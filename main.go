/*Dots
 *A tool to manage dotfiles.
 */
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Dotter interface {
	Getdots(string) ([]string, error)
}

type iu struct{}

func (d *iu) Getdots(wd string) ([]string, error) {
	filenames := []string{}
	files, err := ioutil.ReadDir(wd)
	if err != nil {
		log.Println(err)
	}

	for _, file := range files {
		if file.IsDir() {
			fs, err := d.Getdots(filepath.Join(wd, file.Name()))
			if err != nil {
				log.Println(err)
				continue
			}
			filenames = append(filenames, fs...)

		}
		filenames = append(filenames, filepath.Join(wd, file.Name()))
	}

	return filenames, err
}

type glob struct {
	pattern string
}

func (d *glob) Getdots(wd string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(wd, d.pattern))
	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			log.Println(err)
			continue
		}
		dir := fi.IsDir()
		if !dir {
			continue
		}
		fs, err := d.Getdots(file)
		if err != nil {
			log.Println(err)
			continue
		}
		files = append(files, fs...)
	}
	return files, err
}

func main() {
	wd := filepath.Join(os.Getenv("HOME"), ".dotfiles")
	args := os.Args

	//d := &iu{}
	d := &glob{
		pattern: "*.path",
	}

	files, err := d.Getdots(wd)
	if err != nil {
		log.Fatal(err)
	}
	if len(args) > 1 && args[1] == "link" {
		if len(args) > 2 {
			Link(wd, args[2])
		} else {
			Run(&real{}, files)
		}
	} else {
		Run(&plan{}, files)
	}
}

func Link(wd string, current string) {
	dotfile := filepath.Join(wd, filepath.Base(current))
	pathfile := fmt.Sprintf("%s.%s", dotfile, "path")
	ioutil.WriteFile(pathfile, []byte(current), 0644)
	os.Rename(current, filepath.Join(wd, filepath.Base(current)))
	Run(&real{}, []string{pathfile})
}

type linker interface {
	Symlink(string, string) error
}

func Run(l linker, files []string) {
	for _, file := range files {
		fmt.Println(file)
		p, err := ioutil.ReadFile(file)
		if err != nil {
			log.Println(err)
			continue
		}
		path := os.ExpandEnv(string(p))
		log.Println(path)
		if err := l.Symlink(strings.TrimSuffix(file, ".path"), strings.TrimSpace(path)); err != nil {
			log.Println(err)
		}
	}
}

type real struct{}

func (r *real) Symlink(old string, link string) error {
	return os.Symlink(old, link)
}

type plan struct{}

func (p *plan) Symlink(old string, link string) error {
	fmt.Printf("%s -> %s \n", old, link)
	return nil
}
