package service

// func DeleteDeletedByURL(ctx context.Context, sourceURL string) error {
// 	// _, err := dao.DeleteDeletedByURL(ctx, sourceURL)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// return nil
// 	return database.Default().CancelDeletedByURL(ctx, sourceURL)
// }

// func CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
// 	// return dao.CheckDeletedByURL(ctx, sourceURL)
// 	return database.Default().CheckDeletedByURL(ctx, sourceURL)
// }

// func GetDeletedByURL(ctx context.Context, sourceURL string) (*entity.DeletedRecord, error) {
// 	// return dao.GetDeletedByURL(ctx, sourceURL)
// 	deleted, err := database.Default().GetDeletedByURL(ctx, sourceURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return deleted, nil
// }
