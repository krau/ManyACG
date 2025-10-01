package service

// func CreateLike(ctx context.Context, userID, artworkID primitive.ObjectID) error {
// 	likeModel, err := dao.GetLike(ctx, userID, artworkID)
// 	if errors.Is(err, mongo.ErrNoDocuments) {
// 		like := &types.LikeModel{
// 			UserID:    userID,
// 			ArtworkID: artworkID,
// 		}
// 		_, err := dao.CreateLike(ctx, like)
// 		if err != nil {
// 			return err
// 		}
// 		_, err = dao.IncrementArtworkLikeCountByID(ctx, artworkID)
// 		return err
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	if likeModel != nil {
// 		return errs.ErrLikeExists
// 	}
// 	return nil
// }

// func GetLike(ctx context.Context, userID, artworkID primitive.ObjectID) (*types.LikeModel, error) {
// 	return dao.GetLike(ctx, userID, artworkID)
// }
