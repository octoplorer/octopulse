package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type AssetWrite struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Base64      string `json:"base64" maxLength:"5592416"`
}
type AssetURL struct {
	URL string `json:"url"`
}

func (s *Server) registerAssets() {
	huma.Register(
		s.API,
		huma.Operation{
			OperationID:  "uploadAsset",
			Method:       "POST",
			Path:         "/api/v1/assets",
			MaxBodyBytes: 8 << 20,
		},
		func(ctx context.Context, in *CreateInput[AssetWrite]) (*Output[AssetURL], error) {
			data, e := base64.StdEncoding.DecodeString(in.Body.Base64)
			if e != nil || len(data) > 4<<20 || len(data) == 0 {
				return nil, huma.Error422UnprocessableEntity("Image must be valid base64 and at most 4 MiB")
			}
			dimensions, format, e := image.DecodeConfig(bytes.NewReader(data))
			extensions := map[string]string{"png": ".png", "jpeg": ".jpg", "gif": ".gif"}
			if e != nil || extensions[format] == "" || dimensions.Width < 1 || dimensions.Height < 1 || int64(
				dimensions.Width,
			)*int64(dimensions.Height) > 20000000 {
				return nil, huma.Error422UnprocessableEntity("Upload a PNG, JPEG or GIF image with at most 20 million pixels")
			}
			dir := filepath.Join(s.Config.DataDir, "uploads")
			if e = os.MkdirAll(dir, 0700); e != nil {
				return nil, apiError(ctx, e)
			}
			id := domain.ID()
			name := id + extensions[format]
			path := filepath.Join(dir, name)
			if e = os.WriteFile(path, data, 0600); e != nil {
				return nil, apiError(ctx, e)
			}
			result := AssetURL{URL: "/assets/uploads/" + name}
			e = s.Store.WithTx(ctx, func(t *store.Tx) error {
				if e := t.Put(
					ctx,
					"assets",
					id,
					result,
				); e != nil {
					return e
				}
				return audit(
					ctx,
					t,
					"upload",
					"assets",
					id,
				)
			})
			if e != nil {
				os.Remove(path)
				return nil, apiError(ctx, e)
			}
			return &Output[AssetURL]{Body: result}, nil
		},
	)
}
