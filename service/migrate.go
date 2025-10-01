package service

// func ProcessPicturesHashAndSizeAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
// 	pictures, err := dao.GetNoHashPictures(ctx)
// 	sendMessage := bot != nil && message != nil
// 	if err != nil {
// 		common.Logger.Errorf("Failed to get not processed pictures: %v", err)
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to get not processed pictures: %s",
// 				err.Error(),
// 			))
// 		}
// 		return
// 	}
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Found %d not processed pictures",
// 			len(pictures),
// 		))
// 	}

// 	failed := 0
// 	for _, picture := range pictures {
// 		if err := ProcessPictureHashAndUpdate(ctx, picture.ToPicture()); err != nil {
// 			common.Logger.Errorf("Failed to process picture hash and size: %v", err)
// 			failed++
// 		}
// 	}
// 	common.Logger.Infof("Processed %d pictures, %d failed", len(pictures)-failed, failed)
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Processed %d pictures, %d failed",
// 			len(pictures)-failed,
// 			failed,
// 		))
// 	}
// }

// func StoragePictureRegularAndThumbAndUpdate(ctx context.Context, picture *types.PictureModel) error {
// 	pictureModel, err := dao.GetPictureByID(ctx, picture.ID)
// 	if err != nil {
// 		return err
// 	}
// 	artwork, err := GetArtworkByID(ctx, pictureModel.ArtworkID)
// 	if err != nil {
// 		return err
// 	}
// 	session, err := dao.Client.StartSession()
// 	if err != nil {
// 		return err
// 	}
// 	defer session.EndSession(ctx)
// 	migrateDir := config.Cfg.Storage.CacheDir + "/migrate/"
// 	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
// 		fileBytes, err := storage.GetFile(ctx, picture.StorageInfo.Original)
// 		if err != nil {
// 			return nil, err
// 		}
// 		originalPath := migrateDir + filepath.Base(picture.StorageInfo.Original.Path)
// 		if err := common.MkFile(originalPath, fileBytes); err != nil {
// 			return nil, err
// 		}
// 		defer func() {
// 			common.PurgeFile(originalPath)
// 		}()

// 		regularPath := migrateDir + picture.ID.Hex() + "_regular.webp"
// 		if err := imgtool.CompressImageByFFmpeg(originalPath, regularPath, types.RegularPhotoSideLength); err != nil {
// 			return nil, err
// 		}
// 		defer func() {
// 			common.PurgeFile(regularPath)
// 		}()
// 		basePath := fmt.Sprintf("%s/%s", artwork.SourceType, artwork.Artist.UID)
// 		regularStoragePath := fmt.Sprintf("/regular/%s/%s", basePath, picture.ID.Hex()+"_regular.webp")
// 		regularDetail, err := storage.Save(ctx, regularPath, regularStoragePath, types.StorageType(config.Cfg.Storage.RegularType))
// 		if err != nil {
// 			return nil, err
// 		}
// 		pictureModel.StorageInfo.Regular = regularDetail
// 		if _, err := dao.UpdatePictureStorageInfoByID(ctx, pictureModel.ID, pictureModel.StorageInfo); err != nil {
// 			return nil, err
// 		}

// 		thumbPath := migrateDir + picture.ID.Hex() + "_thumb.webp"
// 		if err := imgtool.CompressImageByFFmpeg(originalPath, thumbPath, types.ThumbPhotoSideLength); err != nil {
// 			return nil, err
// 		}
// 		defer func() {
// 			common.PurgeFile(thumbPath)
// 		}()
// 		thumbStoragePath := fmt.Sprintf("/thumb/%s/%s", basePath, picture.ID.Hex()+"_thumb.webp")
// 		thumbDetail, err := storage.Save(ctx, thumbPath, thumbStoragePath, types.StorageType(config.Cfg.Storage.ThumbType))
// 		if err != nil {
// 			return nil, err
// 		}
// 		pictureModel.StorageInfo.Thumb = thumbDetail
// 		if _, err := dao.UpdatePictureStorageInfoByID(ctx, pictureModel.ID, pictureModel.StorageInfo); err != nil {
// 			return nil, err
// 		}
// 		return nil, nil
// 	}, options.Transaction().SetReadPreference(readpref.Primary()))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func StoragePicturesRegularAndThumbAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
// 	pictures, err := dao.GetNoRegularAndThumbPictures(ctx)
// 	sendMessage := bot != nil && message != nil
// 	if err != nil {
// 		common.Logger.Errorf("Failed to get no regular and thumb pictures: %v", err)
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to get no regular and thumb pictures: %s",
// 				err.Error(),
// 			))
// 		}
// 		return
// 	}
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Found %d no regular and thumb pictures",
// 			len(pictures),
// 		))
// 	}

// 	failed := 0
// 	for _, picture := range pictures {
// 		if err := StoragePictureRegularAndThumbAndUpdate(ctx, picture); err != nil {
// 			common.Logger.Errorf("Failed to storage regular and thumb picture: %v", err)
// 			failed++
// 		}
// 	}
// 	common.Logger.Infof("Processed %d pictures, %d failed", len(pictures)-failed, failed)
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Processed %d pictures, %d failed",
// 			len(pictures)-failed,
// 			failed,
// 		))
// 	}
// }

// func FixTwitterArtists(ctx context.Context, bot *telego.Bot, message *telego.Message) {
// 	client := req.C().ImpersonateChrome().SetCommonRetryCount(3).
// 		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).
// 		SetCommonRetryFixedInterval(2 * time.Second).
// 		EnableDebugLog()
// 	if config.Cfg.Source.Proxy != "" {
// 		client.SetProxyURL(config.Cfg.Source.Proxy)
// 	}
// 	sendMessage := bot != nil && message != nil

// 	collection := dao.DB.Collection("Artists")
// 	if collection == nil {
// 		common.Logger.Errorf("Failed to get collection")
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to get collection",
// 			))
// 		}
// 		return
// 	}
// 	total, err := collection.CountDocuments(ctx, bson.M{"type": "twitter"})
// 	if err != nil {
// 		common.Logger.Errorf("Failed to count artists: %v", err)
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to count artists: %s",
// 				err.Error(),
// 			))
// 		}
// 		return
// 	}
// 	common.Logger.Infof("Found %d artists", total)
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Found %d artists",
// 			total,
// 		))
// 	}
// 	cursor, err := collection.Find(ctx, bson.M{"type": "twitter"})
// 	if err != nil {
// 		common.Logger.Errorf("Failed to find artists: %v", err)
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to find artists: %s",
// 				err.Error(),
// 			))
// 		}
// 		return
// 	}
// 	defer cursor.Close(ctx)
// 	apiBase := fmt.Sprintf("https://api.%s/", config.Cfg.Source.Twitter.FxTwitterDomain)
// 	type ArtistResp struct {
// 		Code    int    `json:"code"`
// 		Message string `json:"message"`
// 		User    struct {
// 			ID         string `json:"id"`
// 			Name       string `json:"name"`
// 			ScreenName string `json:"screen_name"`
// 		} `json:"user"`
// 	}

// 	// 创建一个集合，用于存储失败文档id
// 	if err := dao.DB.CreateCollection(ctx, "failedArtists"); err != nil {
// 		common.Logger.Errorf("Failed to create collection: %v", err)
// 	}
// 	failedCollection := dao.DB.Collection("failedArtists")
// 	failed, count := 0, 0
// 	for cursor.Next(ctx) {
// 		count++
// 		var artist types.ArtistModel
// 		if err := cursor.Decode(&artist); err != nil {
// 			common.Logger.Errorf("Failed to decode artist: %v", err)
// 			failed++
// 			continue
// 		}
// 		updateArtist := func() error {
// 			resp, err := client.R().Get(apiBase + artist.Username)
// 			if err != nil {
// 				common.Logger.Errorf("Failed to get artist: %v", err)
// 				return err
// 			}
// 			var artistResp ArtistResp
// 			if err := json.Unmarshal(resp.Bytes(), &artistResp); err != nil {
// 				common.Logger.Errorf("Failed to unmarshal artist: %v", err)
// 				return err
// 			}
// 			if artistResp.Code != 200 {
// 				common.Logger.Errorf("Failed to get artist: %v", artistResp.Message)
// 				return fmt.Errorf("failed to get artist: %s", artistResp.Message)
// 			}
// 			artist.UID = artistResp.User.ID
// 			artist.Name = artistResp.User.Name
// 			if _, err := dao.UpdateArtist(ctx, &artist); err != nil {
// 				common.Logger.Errorf("Failed to update artist: %v", err)
// 				return err
// 			}
// 			return nil
// 		}
// 		if err := updateArtist(); err != nil {
// 			if _, err := failedCollection.InsertOne(ctx, artist); err != nil {
// 				common.Logger.Errorf("Failed to insert failed artist: %v", err)
// 			}
// 			failed++
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// 	common.Logger.Infof("Processed %d artists, %d failed", count, failed)
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Processed %d artists, %d failed",
// 			count,
// 			failed,
// 		))
// 	}
// }

// func PredictAllArtworkTagsAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
// 	cursor, err := dao.GetCollection("Artworks").Find(ctx, bson.M{})
// 	sendMessage := bot != nil && message != nil
// 	if err != nil {
// 		common.Logger.Errorf("Failed to find artworks: %v", err)
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to find artworks: %s",
// 				err.Error(),
// 			))
// 		}
// 		return
// 	}
// 	defer cursor.Close(ctx)
// 	total, err := dao.GetArtworkCount(ctx, types.R18TypeAll)
// 	if err != nil {
// 		common.Logger.Errorf("Failed to count artworks: %v", err)
// 		if sendMessage {
// 			bot.SendMessage(ctx, telegoutil.Messagef(
// 				message.Chat.ChatID(),
// 				"Failed to count artworks: %s",
// 				err.Error(),
// 			))
// 		}
// 		return
// 	}
// 	failed, count := 0, 0
// 	for cursor.Next(ctx) {
// 		count++
// 		var artwork types.ArtworkModel
// 		if err := cursor.Decode(&artwork); err != nil {
// 			common.Logger.Errorf("Failed to decode artwork: %v", err)
// 			failed++
// 			continue
// 		}
// 		if err := PredictArtworkTagsByIDAndUpdate(ctx, artwork.ID, nil); err != nil {
// 			common.Logger.Errorf("Failed to predict artwork tags: %v", err)
// 			failed++
// 		}
// 	}
// 	common.Logger.Infof("Total %d artworks, processed %d, failed %d", total, count, failed)
// 	if sendMessage {
// 		bot.SendMessage(ctx, telegoutil.Messagef(
// 			message.Chat.ChatID(),
// 			"Total %d artworks, processed %d, failed %d",
// 			total,
// 			count,
// 			failed,
// 		))
// 	}
// }