package form

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/tidwall/gjson"
	"github.com/wfrs-dev/sic/internal/http"
	"github.com/wfrs-dev/sic/internal/model"
	"github.com/wfrs-dev/sic/internal/tui"
)

type titem struct {
	value  string
	label  string
	active bool
}

var theme = huh.ThemeBase16()

type titems []titem

type ProjectForm struct {
	dependencies model.Dependencies
	firstResult  gjson.Result
}

func New() (*ProjectForm, error) {
	f := &ProjectForm{}
	err := f.loadDependencies()
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (f *ProjectForm) loadDependencies() error {
	var err error
	f.firstResult, err = http.FirstRequest()
	if err != nil {
		return fmt.Errorf("error loading dependencies: %w", err)
	}

	values := f.firstResult.Get("dependencies.values")
	var dependencies []model.Dependency
	values.ForEach(func(_, value gjson.Result) bool {
		typo := value.Get("name").String()
		value.Get("values").ForEach(func(_, v gjson.Result) bool {
			dependencies = append(dependencies, model.Dependency{
				ID:          v.Get("id").String(),
				Type:        typo,
				Name:        v.Get("name").String(),
				Description: v.Get("description").String(),
			})

			return true
		})

		return true
	})

	slices.SortFunc(dependencies, func(a, b model.Dependency) int {
		return cmp.Compare(a.ID, b.ID)
	})

	f.dependencies = dependencies

	return nil
}

func (f *ProjectForm) Run() error {
	var req = model.ProjectRequest{}
	var action string

	req.Name = Input("Project name", "")
	fmt.Println(tui.Colorize("<blue+bold>%s  Project name:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Name))

	req.Description = Input("Project description", "")
	fmt.Println(
		tui.Colorize("<blue+bold>%s  Project description:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Description),
	)

	ssel := Select(
		"Project type",
		f.firstResult.Get("type.default").String(),
		f.firstResult.Get("type.type").String(),
		f.firstResult.Get("type.values"),
	)
	parts := strings.SplitN(ssel, "|", 2)
	if len(parts) == 2 {
		action = parts[1]
		req.Project = parts[0]
	}
	fmt.Println(tui.Colorize("<blue+bold>%s  Project type:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Project))

	req.Language = Select(
		"Project language",
		f.firstResult.Get("language.default").String(),
		f.firstResult.Get("language.type").String(),
		f.firstResult.Get("language.values"),
	)
	fmt.Println(tui.Colorize("<blue+bold>%s  Project language:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Language))

	req.SpringBoot = Select(
		"Spring Boot version",
		f.firstResult.Get("bootVersion.default").String(),
		f.firstResult.Get("bootVersion.type").String(),
		f.firstResult.Get("bootVersion.values"),
	)
	fmt.Println(
		tui.Colorize("<blue+bold>%s  Spring Boot version:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.SpringBoot),
	)

	req.Packaging = Select(
		"Project Packaging:",
		f.firstResult.Get("packaging.default").String(),
		f.firstResult.Get("packaging.type").String(),
		f.firstResult.Get("packaging.values"),
	)
	fmt.Println(
		tui.Colorize("<blue+bold>%s  Project packaging:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Packaging),
	)

	req.JavaVersion = Select(
		"Java version (Verify the version installed on your computer)",
		"",
		f.firstResult.Get("javaVersion.type").String(),
		f.firstResult.Get("javaVersion.values"),
	)
	fmt.Println(tui.Colorize("<blue+bold>%s  Java version:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.JavaVersion))

	req.Group = Input("Group", f.firstResult.Get("groupId.default").String())
	fmt.Println(tui.Colorize("<blue+bold>%s  Group:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Group))

	req.Artifact = Input("Artifact", f.firstResult.Get("artifactId.default").String())
	fmt.Println(tui.Colorize("<blue+bold>%s  Artifact:</> %s", tui.Nerdfont("OctDiffAdded", "#"), req.Artifact))

	req.PackageName = fmt.Sprintf("%s.%s", req.Group, req.Artifact)
	fmt.Println(
		tui.Colorize(
			"<blue+bold>%s  Package:</> %s",
			tui.Nerdfont("FaCogs", "#"),
			req.PackageName,
		),
	)

	req.Dependencies = SelectDependencies(f.dependencies)

	fmt.Println()
	if Confirm(tui.Colorize("<yellow+bold>☑ Continue?</>")) {
		fmt.Println(tui.Colorize("<yellow+italic>%s  Building...</>", tui.Nerdfont("FaCog", "-")))
		err := http.CreateProject(req, action)
		if err != nil {
			fmt.Println(tui.Colorize("<red+bold>%s  Error: %s</>", tui.Nerdfont("FaWarning", "!!")))
			os.Exit(1)
		}

		fmt.Println(tui.Colorize("<green+bold>%s  Project created successfully</>", tui.Nerdfont("FaCheck", "✓ ")))
	} else {
		fmt.Println(tui.Colorize("<yellow+italic>%s  Cancelando proyecto</>", tui.Nerdfont("FaWarning", "!")))
	}

	return nil
}

func Input(label, defval string) string {
	var input string
	err := huh.NewInput().
		Title(label).
		Placeholder(defval).
		Value(&input).
		Prompt("=> ").
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("Input cannot be empty")
			}
			return nil
		}).
		WithTheme(theme).
		Run()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return strings.TrimSpace(input)
}

func Select(slabel, defval, typo string, values gjson.Result) string {
	var input string
	var opts titems = make([]titem, 0)
	var id, value, label string
	values.ForEach(func(_, val gjson.Result) bool {
		switch typo {
		case "single-select":
			value = strings.TrimSuffix(val.Get("id").String(), ".RELEASE")
			id = value
			label = val.Get("name").String()
		case "action":
			value = fmt.Sprintf("%s|%s", val.Get("id").String(), val.Get("action").String())
			id = val.Get("id").String()
			label = fmt.Sprintf("%s: %s", val.Get("name").String(), val.Get("description").String())
		}
		opts = append(opts, titem{
			value:  value,
			label:  label,
			active: id == defval,
		})

		return true
	})

	hopts := make([]huh.Option[string], len(opts))
	for i, opt := range opts {
		hopts[i] = huh.NewOption(opt.label, opt.value).Selected(opt.active)
	}
	err := huh.NewSelect[string]().
		Title(slabel).
		Value(&input).
		Options(hopts...).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("No option selected")
			}
			return nil
		}).
		WithTheme(theme).
		Run()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return strings.TrimSpace(input)
}

func SelectDependencies(deps model.Dependencies) string {
	var input []string
	var opts titems = make([]titem, 0)
	for _, dep := range deps {
		opts = append(opts, titem{
			value: dep.ID,
			label: dep.Name,
		})
	}

	hopts := make([]huh.Option[string], len(opts))
	for i, opt := range opts {
		hopts[i] = huh.NewOption(opt.label, opt.value).Selected(opt.active)
	}

	err := huh.NewMultiSelect[string]().
		Title("Dependencies").
		Description("↑ Up 🞄 ↓ Down 🞄 / Search mode 🞄 ⏎  Show Search items/Submit 🞄 ␣ Select item 🞄 Esc Exit search mode").
		Value(&input).
		Options(hopts...).
		WithTheme(theme).
		WithHeight(10).
		Run()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(tui.Colorize("<blue+bold>%s  Dependencies:</>", tui.Nerdfont("OctApps", "*")))
	if len(input) > 0 {
		for _, dep := range input {
			fmt.Println(
				tui.Colorize(
					"   <blue+bold>%s  %s:</> %s",
					tui.Nerdfont("OctPackageDependents", "-"),
					dep,
					findDepDescription(dep, deps),
				),
			)
		}
	} else {
		fmt.Println(tui.Colorize("  <purple+italic>%s  No dependency selected</>", tui.Nerdfont("FaCircleDot", "-")))
	}

	return strings.Join(input, ",")
}

func findDepDescription(id string, deps model.Dependencies) string {
	for _, dep := range deps {
		if dep.ID == id {
			return fmt.Sprintf("[%s] %s", dep.Type, dep.Description)
		}
	}

	return ""
}

func Confirm(msg string) bool {
	fmt.Print(msg + " (y/N): ")
	var char rune
	fmt.Scanf("%c", &char)

	return char == 'y' || char == 'Y' || char == 's' || char == 'S'
}
