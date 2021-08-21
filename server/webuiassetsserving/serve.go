package webuiassetsserving

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var AssetsBasePath string
var dynamicRouteParamSegment = regexp.MustCompile(`\[.+?\]`)

func Intialize(assetsBasePath string) error {
	var err error

	if len(assetsBasePath) == 0 {
		AssetsBasePath, err = determineDefaultAssetsBasePath()
		return err
	} else {
		AssetsBasePath = assetsBasePath
		return nil
	}
}

func SetupRouter(engine *gin.Engine) error {
	return filepath.WalkDir(AssetsBasePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		route := determineRoute(path, d.Name())
		handler := func(ctx *gin.Context) { serve(ctx, path) }
		engine.GET(route, handler)
		engine.HEAD(route, handler)
		return nil
	})
}

func determineDefaultAssetsBasePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Error determining Sqedule executable's own path: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	candidates := []string{
		filepath.Join("webui", "out"),
		filepath.Join(exeDir, "webui-assets"),
		filepath.Join(exeDir, "..", "share", "sqedule-server", "webui"),
	}

	for _, path := range candidates {
		exists, err := fileExists(path)
		if err != nil {
			return "", err
		}
		if exists {
			return path, nil
		}
	}

	return "", fmt.Errorf("Cannot find web UI assets base path (tried: %s)", strings.Join(candidates, ", "))
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Maps resource filesystem paths to HTTP paths as follows:
//
// /assets-base/foo.html                  -> /foo
// /assets-base/index.html                -> /
// /assets-base/prefix/index.html         -> /prefix
// /assets-base/resource/[id]/[id2].html  -> /resource/:id/:id2
// /assets-base/any-other.extension       -> /any-other.extension
func determineRoute(path string, baseName string) string {
	// /assets-base/foo/bar -> foo/bar
	pathRelativeToBase := strings.TrimPrefix(path, AssetsBasePath+"/")

	if strings.HasSuffix(path, ".html") {
		if baseName == "index.html" {
			return "/" + strings.TrimSuffix(pathRelativeToBase, "index.html")
		} else {
			pathRelativeToBaseNoExt := strings.TrimSuffix(pathRelativeToBase, ".html")
			if strings.IndexByte(path, '[') == -1 {
				return "/" + pathRelativeToBaseNoExt
			} else {
				return "/" + dynamicRouteParamSegment.ReplaceAllStringFunc(pathRelativeToBaseNoExt,
					func(paramSegment string) string {
						paramName := strings.TrimSuffix(strings.TrimPrefix(paramSegment, "["), "]")
						return ":" + paramName
					})
			}
		}
	} else {
		return "/" + pathRelativeToBase
	}
}

func serve(ctx *gin.Context, assetPath string) {
	f, err := os.Open(assetPath)
	if err != nil {
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		ctx.String(500, "Error opening file: %s", err.Error())
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		ctx.String(500, "Error querying file metadata: %s", err.Error())
		return
	}

	http.ServeContent(ctx.Writer, ctx.Request, d.Name(), d.ModTime(), f)
}
