package template

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
)

type Options struct {
	// StrictValidators if enabled forces us to use the validator function
	// for a template only if available. Otherwise, the template will be
	// ignored.
	StrictValidators  bool
	UseFilenamePrefix bool
	Plugin            *protogen.Plugin
	Files             embed.FS        `validate:"required"`
	Context           TemplateContext `validate:"required"`
	HelperFunctions   map[string]interface{}
}

// TemplateContext is an interface that a template file context, i.e., the
// object manipulated inside the template file, must implement.
type TemplateContext interface {
	ValidateForExecute() map[string]TemplateValidator
	Extension() string
}

type TemplateValidator func() bool

// Templates is an object that holds information related to a group of
// template files, allowing them to be parsed later.
type Templates struct {
	strictValidators  bool
	useFilenamePrefix bool
	path              string
	prefix            string
	context           TemplateContext
	templates         []*Info
}

type Info struct {
	templateFilename string
	data             []byte
	api              map[string]interface{}
}

// Generated holds the template content already parsed, ready to be saved.
type Generated struct {
	Filename     string
	TemplateName string
	Extension    string
	Data         *bytes.Buffer
}

func (t *Templates) Execute() ([]*Generated, error) {
	var gen []*Generated

	for _, template := range t.templates {
		validator, ok := t.context.ValidateForExecute()[template.templateFilename]
		if !ok && t.strictValidators {
			// The validator should be executed in this case, since we don't
			// have one for this template, we can skip it.
			continue
		}
		if ok && !validator() {
			// Ignores the template if its validation condition is not
			// satisfied
			continue
		}

		tpl, err := parse(template.templateFilename, template.data, template.api)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		if err := tpl.Execute(w, t.context); err != nil {
			return nil, err
		}

		w.Flush()
		gen = append(gen, &Generated{
			Data:         &buf,
			Filename:     t.buildFilename(template.templateFilename),
			TemplateName: template.templateFilename,
			Extension:    t.context.Extension(),
		})
	}

	return gen, nil
}

func (t *Templates) buildFilename(templateFilename string) string {
	filename := templateFilename

	if t.path != "" {
		filename = fmt.Sprintf("%s/%s", t.path, templateFilename)
		if t.useFilenamePrefix {
			filename = fmt.Sprintf("%s/%s.%s", t.path, t.prefix, templateFilename)
		}
	}
	if t.context.Extension() != "" {
		filename += fmt.Sprintf(".%s", t.context.Extension())
	}

	return filename
}

// LoadTemplates loads template files according the desired options and put them
// ready to be executed.
func LoadTemplates(options *Options) (*Templates, error) {
	validate := validator.New()
	if err := validate.Struct(options); err != nil {
		return nil, err
	}

	var (
		filename string
		path     string
	)

	if options.Plugin != nil {
		var err error

		filename, path, err = GetPackageNameAndPath(options.Plugin)
		if err != nil {
			return nil, err
		}
	}

	templates, err := options.Files.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var (
		helperApi = buildDefaultHelperApi()
		tpls      []*Info
	)

	for _, t := range templates {
		data, err := options.Files.ReadFile(t.Name())
		if err != nil {
			return nil, err
		}

		basename := filenameWithoutExtension(t.Name())
		helperApi["templateName"] = func() string {
			return basename
		}

		for k, v := range options.HelperFunctions {
			helperApi[k] = v
		}

		tpls = append(tpls, &Info{
			templateFilename: basename,
			data:             data,
			api:              helperApi,
		})
	}

	return &Templates{
		templates:         tpls,
		path:              path,
		prefix:            filename,
		context:           options.Context,
		strictValidators:  options.StrictValidators,
		useFilenamePrefix: options.UseFilenamePrefix,
	}, nil
}

func buildDefaultHelperApi() map[string]interface{} {
	return template.FuncMap{
		"toLowerCamelCase": strcase.ToLowerCamel,
		"firstLower": func(s string) string {
			c := s[0]
			return strings.ToLower(string(c))
		},
		"toSnake": strcase.ToSnake,
		"toCamel": strcase.ToCamel,
		"toKebab": strcase.ToKebab,
		"trimSuffix": func(s, suffix string) string {
			return strings.TrimSuffix(s, suffix)
		},
	}
}

func parse(key string, data []byte, helperApi template.FuncMap) (*template.Template, error) {
	t, err := template.New(key).Funcs(helperApi).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func filenameWithoutExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

// GetPackageNameAndPath try to retrieve the golang module name from the list of .proto
// files.
func GetPackageNameAndPath(plugin *protogen.Plugin) (string, string, error) {
	if len(plugin.Files) == 0 {
		return "", "", errors.New("cannot find the module name without .proto files")
	}

	// The last file in the slice is always the main .proto file that is being
	// "compiled" by protoc.
	file := plugin.Files[len(plugin.Files)-1]

	return string(file.GoPackageName), filepath.Dir(file.GeneratedFilenamePrefix), nil
}
