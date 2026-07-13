package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/imseek"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/tagging"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/spf13/cobra"
)

var imseekCmd = &cobra.Command{
	Use:   "imseek",
	Short: "Manage local feature image index",
}

var (
	imseekAddLimit   int
	imseekAddOffset  int
	imseekTrainNList int
	imseekTrainSamp  int
	imseekTrainForce bool
)

var imseekAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Extract features for pictures in the database",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImseekAdd(cmd.Context()))
	},
}

var imseekTrainCmd = &cobra.Command{
	Use:   "train",
	Short: "Train the IVF quantizer",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImseekTrain(cmd.Context()))
	},
}

var imseekBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the IVF inverted index",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImseekBuild(cmd.Context()))
	},
}

var imseekRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Re-index all pictures, then train and build",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImseekRebuild(cmd.Context()))
	},
}

var imseekStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show feature index status",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImseekStatus(cmd.Context()))
	},
}

func init() {
	rootCmd.AddCommand(imseekCmd)
	imseekCmd.AddCommand(
		imseekAddCmd,
		imseekTrainCmd,
		imseekBuildCmd,
		imseekRebuildCmd,
		imseekStatusCmd,
	)

	imseekAddCmd.Flags().IntVarP(&imseekAddLimit, "limit", "n", 0, "Max pictures to process (0=all)")
	imseekAddCmd.Flags().IntVar(&imseekAddOffset, "offset", 0, "Artwork pagination offset")

	imseekTrainCmd.Flags().IntVarP(&imseekTrainNList, "nlist", "l", 0, "Number of clusters (0=auto)")
	imseekTrainCmd.Flags().IntVarP(&imseekTrainSamp, "samples", "s", 0, "Training samples (0=auto)")
	imseekTrainCmd.Flags().BoolVarP(&imseekTrainForce, "force", "f", false, "Retrain even if quantizer exists")
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type imseekRuntime struct {
	cfg  runtimecfg.Config
	eng  imseek.Engine
	serv *service.Service
	stop func() error
}

func openImseekRuntime(ctx context.Context, autoBuild bool) (*imseekRuntime, error) {
	cfg := runtimecfg.Get()
	if !cfg.Imseek.Enable {
		return nil, fmt.Errorf("imseek is disabled: set [imseek] enable = true in config.toml")
	}
	log.SetDefault(log.New(log.Config{LogFile: cfg.Log.FilePath, MaxBackups: int(cfg.Log.BackupNum)}))

	closer, err := infra.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}

	eng, err := imseek.Init(ctx, imseek.Config{
		Enable:           true,
		DataDir:          cfg.Imseek.DataDir,
		Distance:         cfg.Imseek.Distance,
		Count:            cfg.Imseek.Count,
		K:                cfg.Imseek.K,
		NProbe:           cfg.Imseek.NProbe,
		NFeatures:        cfg.Imseek.NFeatures,
		MaxHeight:        cfg.Imseek.MaxHeight,
		MaxWidth:         cfg.Imseek.MaxWidth,
		AutoBuild:        autoBuild,
		BuildDebounceSec: cfg.Imseek.BuildDebounceSec,
		MinMatches:       cfg.Imseek.MinMatches,
		MinScore:         cfg.Imseek.MinScore,
	})
	if err != nil {
		_ = closer()
		return nil, err
	}

	serv := service.NewService(
		repo.Repositories(database.Default()),
		nil,
		tagging.Default(),
		storage.Storages(),
		source.Sources(),
		cfg.Storage,
		service.WithImseek(eng),
	)

	return &imseekRuntime{
		cfg:  cfg,
		eng:  eng,
		serv: serv,
		stop: func() error {
			_ = eng.Close()
			return closer()
		},
	}, nil
}

// indexAllPictures walks artworks and indexes each picture (ORB → imseek.db).
func indexAllPictures(ctx context.Context, serv *service.Service, limit, offset int) (indexed, failed int, err error) {
	const page = 50
	t0 := time.Now()
	artworkOffset := offset
	for {
		aws, err := serv.QueryArtworks(ctx, query.ArtworksDB{
			Paginate: query.Paginate{Limit: page, Offset: artworkOffset},
		})
		if err != nil {
			return indexed, failed, err
		}
		if len(aws) == 0 {
			break
		}
		for _, aw := range aws {
			for _, pic := range aw.Pictures {
				if limit > 0 && indexed+failed >= limit {
					log.Info("imseek add reached limit", "limit", limit, "indexed", indexed, "failed", failed)
					return indexed, failed, nil
				}
				if err := serv.IndexPictureForImseek(ctx, pic); err != nil {
					log.Warn("index picture failed", "id", pic.ID, "err", err)
					failed++
					continue
				}
				indexed++
				if indexed%100 == 0 {
					log.Info("imseek add progress", "indexed", indexed, "failed", failed, "dur", time.Since(t0))
				}
			}
		}
		artworkOffset += len(aws)
		if len(aws) < page {
			break
		}
	}
	return indexed, failed, nil
}

func runImseekAdd(ctx context.Context) error {
	rt, err := openImseekRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()

	t0 := time.Now()
	indexed, failed, err := indexAllPictures(ctx, rt.serv, imseekAddLimit, imseekAddOffset)
	if err != nil {
		return err
	}
	log.Info("imseek add complete", "indexed", indexed, "failed", failed, "dur", time.Since(t0))
	st, _ := rt.eng.Status(ctx)
	log.Info("imseek status", "images", st.Images, "vectors", st.Vectors, "unindexed", st.Unindexed, "trained", st.Trained)
	return nil
}

func runImseekTrain(ctx context.Context) error {
	rt, err := openImseekRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()

	return rt.eng.Train(ctx, imseek.TrainOpts{
		NList:   imseekTrainNList,
		Samples: imseekTrainSamp,
		Force:   imseekTrainForce,
	})
}

func runImseekBuild(ctx context.Context) error {
	rt, err := openImseekRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()
	return rt.eng.Build(ctx)
}

func runImseekRebuild(ctx context.Context) error {
	rt, err := openImseekRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()

	t0 := time.Now()
	indexed, failed, err := indexAllPictures(ctx, rt.serv, 0, 0)
	if err != nil {
		return err
	}
	log.Info("indexing done, train+build…", "indexed", indexed, "failed", failed, "dur", time.Since(t0))
	if err := rt.eng.Rebuild(ctx); err != nil {
		return fmt.Errorf("rebuild: %w", err)
	}
	log.Info("imseek rebuild complete", "total_dur", time.Since(t0))
	return nil
}

func runImseekStatus(ctx context.Context) error {
	rt, err := openImseekRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()

	st, err := rt.eng.Status(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("data_dir:    %s\n", st.DataDir)
	fmt.Printf("images:      %d\n", st.Images)
	fmt.Printf("vectors:     %d\n", st.Vectors)
	fmt.Printf("unindexed:   %d\n", st.Unindexed)
	fmt.Printf("trained:     %v\n", st.Trained)
	fmt.Printf("searcher:    %v\n", st.SearcherOK)
	return nil
}
