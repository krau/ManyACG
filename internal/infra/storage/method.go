package storage

// // 保存图片的所有尺寸
// func SaveAll(ctx context.Context, artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error) {
// 	if len(Storages) == 0 {
// 		return &types.StorageInfo{
// 			Original: nil,
// 			Regular:  nil,
// 			Thumb:    nil,
// 		}, ErrNoStorages
// 	}
// 	common.Logger.Infof("saving picture %d of artwork %s", picture.Index, artwork.Title)
// 	originalBytes, err := common.DownloadWithCache(ctx, picture.Original, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	mimeType := mimetype.Detect(originalBytes)

// 	filePath := filepath.Join(config.Get().Storage.CacheDir, common.MD5Hash(picture.Original)) + mimeType.Extension()
// 	if err := common.MkFile(filePath, originalBytes); err != nil {
// 		return nil, err
// 	}
// 	defer func() {
// 		go common.RmFileAfter(filePath, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
// 	}()
// 	originalStorageFileName, err := source.GetFileName(artwork, picture)
// 	if err != nil {
// 		return nil, err
// 	}
// 	originalStoragePath := fmt.Sprintf("/%s/%s/%s", artwork.SourceType, artwork.Artist.UID, originalStorageFileName)
// 	originalStorage, ok := Storages[types.StorageType(config.Get().Storage.OriginalType)]
// 	if !ok {
// 		common.Logger.Fatalf("Unknown storage type: %s", config.Get().Storage.OriginalType)
// 		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, config.Get().Storage.OriginalType)
// 	}

// 	originalDetail, err := originalStorage.Save(ctx, filePath, originalStoragePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var regularDetail *types.StorageDetail
// 	if config.Get().Storage.RegularType != "" {
// 		regularStorage, ok := Storages[types.StorageType(config.Get().Storage.RegularType)]
// 		if !ok {
// 			common.Logger.Fatalf("Unknown storage type: %s", config.Get().Storage.RegularType)
// 			return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, config.Get().Storage.RegularType)
// 		}
// 		regularOutputPath := fmt.Sprintf("%s_regular.%s", filePath[:len(filePath)-len(filepath.Ext(filePath))], config.Get().Storage.RegularFormat)
// 		if err := imgtool.CompressImage(filePath, regularOutputPath, config.Get().Storage.RegularFormat, types.RegularPhotoSideLength); err != nil {
// 			return nil, err
// 		}
// 		defer func() {
// 			go common.RmFileAfter(regularOutputPath, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
// 		}()

// 		if picture.ID == "" {
// 			picture.ID = primitive.NewObjectID().Hex()
// 		}
// 		regularStorageFileName := picture.ID + "_regular." + config.Get().Storage.RegularFormat
// 		regularStoragePath := fmt.Sprintf("/regular/%s/%s/%s", artwork.SourceType, artwork.Artist.UID, regularStorageFileName)

// 		regularDetail, err = regularStorage.Save(ctx, regularOutputPath, regularStoragePath)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	var thumbDetail *types.StorageDetail
// 	if config.Get().Storage.ThumbType != "" {
// 		thumbStorage, ok := Storages[types.StorageType(config.Get().Storage.ThumbType)]
// 		if !ok {
// 			common.Logger.Fatalf("Unknown storage type: %s", config.Get().Storage.ThumbType)
// 			return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, config.Get().Storage.ThumbType)
// 		}
// 		thumbOutputPath := fmt.Sprintf("%s_thumb.%s", filePath[:len(filePath)-len(filepath.Ext(filePath))], config.Get().Storage.ThumbFormat)
// 		if err := imgtool.CompressImage(filePath, thumbOutputPath, config.Get().Storage.ThumbFormat, types.ThumbPhotoSideLength); err != nil {
// 			return nil, err
// 		}

// 		defer func() {
// 			go common.RmFileAfter(thumbOutputPath, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
// 		}()

// 		if picture.ID == "" {
// 			picture.ID = primitive.NewObjectID().Hex()
// 		}
// 		thumbStorageFileName := picture.ID + "_thumb." + config.Get().Storage.ThumbFormat
// 		thumbStoragePath := fmt.Sprintf("/thumb/%s/%s/%s", artwork.SourceType, artwork.Artist.UID, thumbStorageFileName)

// 		thumbDetail, err = thumbStorage.Save(ctx, thumbOutputPath, thumbStoragePath)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return &types.StorageInfo{
// 		Original: originalDetail,
// 		Regular:  regularDetail,
// 		Thumb:    thumbDetail,
// 	}, nil
// }

// func Save(ctx context.Context, filePath string, storagePath string, storageType types.StorageType) (*types.StorageDetail, error) {
// 	if storage, ok := Storages[storageType]; ok {
// 		return storage.Save(ctx, filePath, storagePath)
// 	} else {
// 		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, storageType)
// 	}
// }

// var storageLocks sync.Map

// func GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
// 	detail, err := applyRule(detail)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if detail.Type != types.StorageTypeLocal {
// 		lock, _ := storageLocks.LoadOrStore(detail.String(), &sync.Mutex{})
// 		lock.(*sync.Mutex).Lock()
// 		defer func() {
// 			lock.(*sync.Mutex).Unlock()
// 			storageLocks.Delete(detail)
// 		}()
// 	}
// 	if storage, ok := Storages[detail.Type]; ok {
// 		file, err := storage.GetFile(ctx, detail)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return file, nil
// 	} else {
// 		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, detail.Type)
// 	}
// }

// func Delete(ctx context.Context, info *types.StorageDetail) error {
// 	if storage, ok := Storages[info.Type]; ok {
// 		return storage.Delete(ctx, info)
// 	} else {
// 		return fmt.Errorf("%w: %s", errs.ErrStorageUnkown, info.Type)
// 	}
// }

// func DeleteAll(ctx context.Context, info *types.StorageInfo) error {
// 	var wg sync.WaitGroup
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()
// 	errChan := make(chan error)
// 	for _, detail := range []*types.StorageDetail{info.Original, info.Regular, info.Thumb} {
// 		if detail == nil {
// 			continue
// 		}
// 		wg.Add(1)
// 		go func(detail *types.StorageDetail) {
// 			defer wg.Done()
// 			if err := Delete(ctx, detail); err != nil {
// 				errChan <- err
// 				cancel()
// 			}
// 		}(detail)
// 	}
// 	go func() {
// 		wg.Wait()
// 		close(errChan)
// 	}()
// 	for err := range errChan {
// 		return err
// 	}
// 	return nil
// }

// func ServeFile(ctx *gin.Context, detail *types.StorageDetail) {
// 	if detail == nil || detail.Path == "" {
// 		utils.GinErrorResponse(ctx, errors.New("invalid storage detail"), http.StatusInternalServerError, "Invalid storage detail")
// 		return
// 	}
// 	switch detail.Type {
// 	case types.StorageTypeLocal:
// 		ctx.File(detail.Path)
// 	default:
// 		data, err := GetFile(ctx, detail)
// 		if err != nil {
// 			common.Logger.Errorf("Failed to get file: %v", err)
// 			utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
// 			return
// 		}
// 		mimeType := mimetype.Detect(data)
// 		ctx.Data(http.StatusOK, mimeType.String(), data)
// 	}
// }

// func GetFileStream(ctx context.Context, detail *types.StorageDetail) (io.ReadCloser, error) {
// 	if detail == nil {
// 		return nil, errors.New("storage detail is nil")
// 	}
// 	if detail.Type != types.StorageTypeLocal {
// 		lock, _ := storageLocks.LoadOrStore(detail.String(), &sync.Mutex{})
// 		lock.(*sync.Mutex).Lock()
// 		defer func() {
// 			lock.(*sync.Mutex).Unlock()
// 			storageLocks.Delete(detail)
// 		}()
// 	}
// 	if storage, ok := Storages[detail.Type]; ok {
// 		file, err := storage.GetFileStream(ctx, detail)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return file, nil
// 	} else {
// 		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, detail.Type)
// 	}
// }
