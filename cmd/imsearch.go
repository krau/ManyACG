package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/imsearch"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/tagging"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/spf13/cobra"
)

var imsearchCmd = &cobra.Command{
	Use:   "imsearch",
	Short: "Manage local feature image index",
}

var (
	imsearchAddLimit   int
	imsearchAddOffset  int
	imsearchTrainNList int
	imsearchTrainSamp  int
	imsearchTrainForce bool
)

var imsearchAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Extract features for pictures in the database",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImsearchAdd(cmd.Context()))
	},
}

var imsearchTrainCmd = &cobra.Command{
	Use:   "train",
	Short: "Train the IVF quantizer",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImsearchTrain(cmd.Context()))
	},
}

var imsearchBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the IVF inverted index",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImsearchBuild(cmd.Context()))
	},
}

var imsearchRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Re-index all pictures, then train and build",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImsearchRebuild(cmd.Context()))
	},
}

var imsearchStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show feature index status",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnErr(runImsearchStatus(cmd.Context()))
	},
}

func init() {
	rootCmd.AddCommand(imsearchCmd)
	imsearchCmd.AddCommand(
		imsearchAddCmd,
		imsearchTrainCmd,
		imsearchBuildCmd,
		imsearchRebuildCmd,
		imsearchStatusCmd,
	)

	imsearchAddCmd.Flags().IntVarP(&imsearchAddLimit, "limit", "n", 0, "Max pictures to process (0=all)")
	imsearchAddCmd.Flags().IntVar(&imsearchAddOffset, "offset", 0, "Artwork pagination offset")

	imsearchTrainCmd.Flags().IntVarP(&imsearchTrainNList, "nlist", "l", 0, "Number of clusters (0=auto)")
	imsearchTrainCmd.Flags().IntVarP(&imsearchTrainSamp, "samples", "s", 0, "Training samples (0=auto)")
	imsearchTrainCmd.Flags().BoolVarP(&imsearchTrainForce, "force", "f", false, "Retrain even if quantizer exists")
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type imsearchRuntime struct {
	cfg  runtimecfg.Config
	eng  imsearch.Engine
	serv *service.Service
	stop func() error
}

func openImsearchRuntime(ctx context.Context, autoBuild bool) (*imsearchRuntime, error) {
	cfg := runtimecfg.Get()
	if !cfg.Imsearch.Enable {
		return nil, fmt.Errorf("imsearch is disabled: set [imsearch] enable = true in config.toml")
	}
	log.SetDefault(log.New(log.Config{LogFile: cfg.Log.FilePath, MaxBackups: int(cfg.Log.BackupNum)}))

	closer, err := infra.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}

	eng, err := imsearch.Init(ctx, imsearch.Config{
		Enable:           true,
		DataDir:          cfg.Imsearch.DataDir,
		Distance:         cfg.Imsearch.Distance,
		Count:            cfg.Imsearch.Count,
		K:                cfg.Imsearch.K,
		NProbe:           cfg.Imsearch.NProbe,
		NFeatures:        cfg.Imsearch.NFeatures,
		MaxHeight:        cfg.Imsearch.MaxHeight,
		MaxWidth:         cfg.Imsearch.MaxWidth,
		AutoBuild:        autoBuild,
		BuildDebounceSec: cfg.Imsearch.BuildDebounceSec,
		MinMatches:       cfg.Imsearch.MinMatches,
		MinScore:         cfg.Imsearch.MinScore,
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
		service.WithImsearch(eng),
	)

	return &imsearchRuntime{
		cfg:  cfg,
		eng:  eng,
		serv: serv,
		stop: func() error {
			_ = eng.Close()
			return closer()
		},
	}, nil
}

// indexAllPictures walks artworks and indexes each picture (ORB → imsearch.db).
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
					log.Info("imsearch add reached limit", "limit", limit, "indexed", indexed, "failed", failed)
					return indexed, failed, nil
				}
				if err := serv.IndexPictureForImsearch(ctx, pic); err != nil {
					log.Warn("index picture failed", "id", pic.ID, "err", err)
					failed++
					continue
				}
				indexed++
				if indexed%100 == 0 {
					log.Info("imsearch add progress", "indexed", indexed, "failed", failed, "dur", time.Since(t0))
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

func runImsearchAdd(ctx context.Context) error {
	rt, err := openImsearchRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()

	t0 := time.Now()
	indexed, failed, err := indexAllPictures(ctx, rt.serv, imsearchAddLimit, imsearchAddOffset)
	if err != nil {
		return err
	}
	log.Info("imsearch add complete", "indexed", indexed, "failed", failed, "dur", time.Since(t0))
	st, _ := rt.eng.Status(ctx)
	log.Info("imsearch status", "images", st.Images, "vectors", st.Vectors, "unindexed", st.Unindexed, "trained", st.Trained)
	return nil
}

func runImsearchTrain(ctx context.Context) error {
	rt, err := openImsearchRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()

	return rt.eng.Train(ctx, imsearch.TrainOpts{
		NList:   imsearchTrainNList,
		Samples: imsearchTrainSamp,
		Force:   imsearchTrainForce,
	})
}

func runImsearchBuild(ctx context.Context) error {
	rt, err := openImsearchRuntime(ctx, false)
	if err != nil {
		return err
	}
	defer rt.stop()
	return rt.eng.Build(ctx)
}

func runImsearchRebuild(ctx context.Context) error {
	rt, err := openImsearchRuntime(ctx, false)
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
	log.Info("imsearch rebuild complete", "total_dur", time.Since(t0))
	return nil
}

func runImsearchStatus(ctx context.Context) error {
	rt, err := openImsearchRuntime(ctx, false)
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
