// Copyright 2016 The wc-lang-packs-server authors. All rights reserved.
// Use of this source code is governed by a GPL v2.0 license that can be found
// at http://www.gnu.org/licenses/gpl-2.0.html

// wc-lang-packs-server serves the translation API for WooCommerce extension and
// language packs (zip file containing .mo and .po files).
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Automattic/wc-lang-packs-server/locales"
)

// Translation represents translation of a project. Used internally to serve
// WooThemes Helper. This mimics response from https://api.wordpress.org/translations/plugins/1.0/
type Translation struct {
	Language     string `json:"language"`
	LastModified string `json:"last_modified"`
	EnglishName  string `json:"english_name"`
	NativeName   string `json:"native_name"`
	Package      string `json:"package"` // Download URL for LP
}

// Project represents project in GP. Follows the format of /api/projects/{project}
type Project struct {
	TranslationSets []TranslationSet `json:"translation_sets"`
	SubProjects     []SubProject     `json:"sub_projects"`
}

// TranslationSet represents translation set in GP project. Follows the format of
// /api/projects/{project}
type TranslationSet struct {
	Name         string `json:"name"`
	Locale       string `json:"locale"`
	WPLocale     string `json:"wp_locale"`
	LastModified string `json:"last_modified,omitempty"`
}

// Unmarshaller for TranslationSet as last_modified's value could be string or bool.
func (ts *TranslationSet) UnmarshalJSON(data []byte) error {
	var i interface{}
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	m := i.(map[string]interface{})
	for k, v := range m {
		switch k {
		case "name":
			ts.Name = v.(string)
		case "locale":
			ts.Locale = v.(string)
		case "wp_locale":
			ts.WPLocale = v.(string)
		case "last_modfified":
			switch v.(type) {
			case string:
				ts.LastModified = v.(string)
			default:
				ts.LastModified = time.Now().Format("2006-01-02 15:04:05")
			}
		}
	}

	return nil
}

// Project represents sub-project of a project in GP. Follows the format of
// /api/projects/{project}
type SubProject struct {
	Slug   string
	Path   string
	Active string
}

var (
	gpURL         = flag.String("gpURL", "https://translate.wordpress.com/projects/", "Root project of GlotPress")
	gpApiURL      = flag.String("gpApiURL", "https://translate.wordpress.com/api/projects/", "Root API project of GlotPress")
	dbPath        = flag.String("db", os.TempDir()+"wc-lang-packs.db", "Full path to DB file")
	seedDb        = flag.Bool("seed", false, "Seed the DB before serving requests")
	exposeDb      = flag.Bool("exposedb", false, "Expose /_db/ to dump in-memory DB as JSON")
	listenAddr    = flag.String("listen", ":8081", "HTTP listen address")
	mode          = flag.String("mode", "poll", "Check mode, 'poll' or 'notified'")
	pollInterval  = flag.Duration("poll-interval", time.Minute*10, "Interval to poll translate.wordpress.com API if mode is poll")
	postKey       = flag.String("update-key", "my-secret-key", "Key to post update if mode is notified")
	downloadsPath = flag.String("downloads-path", os.TempDir()+"downloads", "Full path to serve language packs files")

	// Translations mapped by project, version, and wp_locale.
	db map[string]map[string]map[string]*Translation

	checker *GPChecker
)

func main() {
	flag.Parse()

	db = make(map[string]map[string]map[string]*Translation)
	if *seedDb {
		seed()
	}

	checker = new(GPChecker)

	switch *mode {
	case "poll":
		go checker.run()
	case "notified":
		http.HandleFunc("/api/v1/update", handleUpdate)
		http.HandleFunc("/api/v1/update/", handleUpdate)
	}

	http.Handle("/api/v1/plugins", jsonContent(http.HandlerFunc(handlePlugins)))
	http.Handle("/api/v1/plugins/", jsonContent(http.HandlerFunc(handlePlugins)))
	http.Handle("/api/v1/themes", jsonContent(http.HandlerFunc(handleThemes)))
	http.Handle("/api/v1/themes/", jsonContent(http.HandlerFunc(handleThemes)))
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir(*downloadsPath))))

	if *exposeDb {
		http.Handle("/_db", jsonContent(http.HandlerFunc(handleDbExposer)))
		http.Handle("/_db/", jsonContent(http.HandlerFunc(handleDbExposer)))
	}

	log.Println("Listening at " + *listenAddr)
	log.Println("Update mode " + *mode)
	log.Println("Serving /downloads/ from " + *downloadsPath)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

// seed seeds the DB.
func seed() {
	log.Println("Seeding the DB")
}

// handleUpdate handles ping from GlotPress whenever a translation is updated.
func handleUpdate(w http.ResponseWriter, r *http.Request) {
	jsonError("error_not_implemented", "Not implemented", w, r)
}

// handlePlugins handle API request for /plugins endpoint.
func handlePlugins(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		jsonError("missing_slug", "Missing slug in query string", w, r)
		return
	}

	ver := r.URL.Query().Get("version")
	if ver == "" {
		jsonError("missing_version", "Missing version in query string", w, r)
		return
	}

	// Optional. If missing returns it as array.
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		if _, ok := db[slug][ver]; !ok {
			jsonError("translation_does_not_exist", fmt.Sprintf("Translation for plugin %s version %s does not exist", slug, ver), w, r)
			return
		}

		if err := json.NewEncoder(w).Encode(db[slug][ver]); err != nil {
			jsonError("error_json_encode", err.Error(), w, r)
			return
		}
	}

	if _, ok := db[slug][ver][locale]; !ok {
		jsonError("translation_does_not_exist", fmt.Sprintf("Translation %s for plugin %s version %s", locale, slug, ver), w, r)
	}
	if err := json.NewEncoder(w).Encode(db[slug][ver][locale]); err != nil {
		jsonError("error_json_encode", err.Error(), w, r)
		return
	}
}

// handleThemes handles API request for /themes endpoint.
func handleThemes(w http.ResponseWriter, r *http.Request) {
	jsonError("error_not_implemented", "Not implemented", w, r)
}

// handleDbExposer handles /_db/ endpoint.
func handleDbExposer(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(&db); err != nil {
		jsonError("error_json_encode", err.Error(), w, r)
	}
}

// jsonContent is middleware to set response's Content-Type to application/json.
func jsonContent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		h.ServeHTTP(w, r)
	})
}

// errorResp represents error response.
type errorResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// jsonError handles error response.
func jsonError(code string, message string, w http.ResponseWriter, r *http.Request) {
	var c int
	switch code {
	case "missing_slug":
		c = http.StatusBadRequest
	case "missing_version":
		c = http.StatusBadRequest
	case "translation_does_not_exist":
		c = http.StatusNotFound
	default:
		c = http.StatusInternalServerError
	}

	w.WriteHeader(c)

	er := errorResp{code, message}
	json.NewEncoder(w).Encode(&er)
}

// GPChecker represents GlotPress checker.
type GPChecker struct {
	err error
}

// run runs in its own goroutine.
func (gpc *GPChecker) run() {
	log.Printf("Start polling each %s", *pollInterval)

	for {
		gpc.poll()
		if gpc.err != nil {
			log.Printf("Error during poll: %s\n", gpc.err)
			gpc.err = nil
		}
		log.Printf("Sleep for %v\n", *pollInterval)
		time.Sleep(*pollInterval)
	}
}

// poll runs from the run loop goroutine.
func (gpc *GPChecker) poll() {
	wc, err := gpc.fetchProject("woocommerce")
	if err != nil {
		gpc.err = err
		return
	}

	// Iterate each WooCommerce projects (e.g., woocommerce-bookings)
	for _, p := range wc.SubProjects {
		// Project doesn't exists in db.
		if _, ok := db[p.Slug]; !ok {
			db[p.Slug] = make(map[string]map[string]*Translation)
		}

		fp, err := gpc.fetchProject(p.Path)
		if err != nil {
			log.Printf("Error in fetching: %v\n", err)
			continue
		}

		// Subprojects are versions of an extension.
		for _, pv := range fp.SubProjects {
			// Fetch extension version.
			v, err := gpc.fetchProject(pv.Path)
			if err != nil {
				log.Printf("Error in fetching: %v\n", err)
				continue
			}

			// Extension's version doesn't exists in db.
			if _, ok := db[p.Slug][pv.Slug]; !ok {
				db[p.Slug][pv.Slug] = make(map[string]*Translation)
			}

			// Iterate translation set.
			for _, ts := range v.TranslationSets {
				var t *Translation
				if _, ok := db[p.Slug][pv.Slug][ts.WPLocale]; !ok {
					t = getTranslationByLocale(ts.WPLocale)
					t.LastModified = ts.LastModified
					t.Package = buildPackageZip(pv.Path, ts.WPLocale)

					db[p.Slug][pv.Slug][ts.WPLocale] = t
				}

				// Only update if last_modified differs
				if db[p.Slug][pv.Slug][ts.WPLocale].LastModified != ts.LastModified {
					t = getTranslationByLocale(ts.WPLocale)
					t.LastModified = ts.LastModified
					t.Package = buildPackageZip(pv.Path, ts.WPLocale)

					db[p.Slug][pv.Slug][ts.WPLocale] = t
				}
			}
		}
	}
}

// fetchProject fetches the project from GlotPress API.
func (gpc *GPChecker) fetchProject(path string) (p *Project, err error) {
	log.Printf("Fetching project at %s\n", getApiURL(path))
	resp, err := http.Get(getApiURL(path))
	defer resp.Body.Close()
	if err != nil {
		return p, err
	}

	if err = json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return p, err
	}

	return p, err
}

// getApiURL returns fullpath of API URL of a given path.
func getApiURL(path string) string {
	return *gpApiURL + path
}

// getPackageURL returns language packs download URL for a given path and language package file.
func getPackageURL(path, lpName string) string {
	return "/downloads/" + path + "/" + lpName
}

// buildPackageZip builds the language pack file, .zip file containing .po and
// .mo files, and returns the URL.
func buildPackageZip(path, locale string) string {
	path = strings.TrimPrefix(path, "woocommerce/")
	parts := strings.Split(path, "/")

	lpName := fmt.Sprintf("%s-%s-%s.zip", parts[0], parts[1], locale)
	log.Println("Building language pack " + lpName)

	pomoDir := filepath.Join(os.TempDir(), filepath.Clean(path), locale)
	if err := os.MkdirAll(pomoDir, 0755); err != nil {
		log.Printf("Error creating directory %s for POMO files\n", pomoDir)
		return ""
	}
	defer os.RemoveAll(pomoDir)

	exportURL := *gpURL + "/" + path + "/" + locale

	po := filepath.Join(pomoDir, parts[0]+"-"+locale+".po")
	if err := downloadTranslation(exportURL+"?format=po", po); err != nil {
		log.Printf("Error downloading .po file: %v\n", err)
		return ""
	}
	defer os.Remove(po)

	mo := filepath.Join(pomoDir, parts[0]+"-"+locale+".mo")
	if err := downloadTranslation(exportURL+"?format=mo", mo); err != nil {
		log.Printf("Error downloading .mo file: %v\n", err)
		return ""
	}
	defer os.Remove(mo)

	lpDir := filepath.Join(*downloadsPath, filepath.Clean(path))
	if err := os.MkdirAll(lpDir, 0755); err != nil {
		log.Printf("Error creating directory for Language Packs: %v\n", err)
		return ""
	}

	if err := zipPOMOFiles(pomoDir, filepath.Join(lpDir, lpName)); err != nil {
		log.Printf("Error zipping th POMO files: %v\n", err)
		return ""
	}

	return getPackageURL(path, lpName)
}

// downloadTranslation downloads the translation file (.po or .mo) from GlotPress
// into dst and returns error if exists.
func downloadTranslation(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// zipPOMOFiles zip the .po and .mo files in src to dst.
func zipPOMOFiles(srcDir, zipDst string) error {
	zipFile, err := os.Create(zipDst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	arch := zip.NewWriter(zipFile)
	defer arch.Close()

	_, err = os.Stat(srcDir)
	if err != nil {
		return err
	}

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name = "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := arch.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(writer, f)
		return err
	})

	return err
}

// getTranslationByLocale gets translation by locale.
func getTranslationByLocale(locale string) (t *Translation) {
	t = new(Translation)
	t.Language = locale
	t.EnglishName = locales.GetLocaleProp(locale, "EnglishName")
	t.NativeName = locales.GetLocaleProp(locale, "NativeName")

	return t
}
