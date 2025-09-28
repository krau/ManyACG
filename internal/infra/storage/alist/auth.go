package alist

// func getJwtToken(ctx context.Context) (string, error) {
// 	resp, err := reqClient.R().SetContext(ctx).SetBodyJsonMarshal(loginReq).Post("/api/auth/login")
// 	if err != nil {
// 		return "", err
// 	}
// 	var loginResp loginResponse
// 	if err := json.Unmarshal(resp.Bytes(), &loginResp); err != nil {
// 		return "", err
// 	}
// 	if loginResp.Code != http.StatusOK {
// 		return "", fmt.Errorf("%w: %s", ErrAlistLoginFailed, loginResp.Message)
// 	}
// 	return loginResp.Data.Token, nil
// }

// func refreshJwtToken(client *req.Client) {
// 	for {
// 		time.Sleep(time.Duration(config.Get().Storage.Alist.TokenExpire) * time.Second)
// 		token, err := getJwtToken(context.Background())
// 		if err != nil {
// 			common.Logger.Errorf("Failed to refresh jwt token: %v", err)
// 			continue
// 		}
// 		client.SetCommonHeader("Authorization", token)
// 		common.Logger.Info("Refreshed Alist jwt token")
// 	}
// }
