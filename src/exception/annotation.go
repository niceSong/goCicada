package exception

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var annotationRegex = "@CicadaError\\([\\s]*code[\\s]*=[\\s]*(.*?),[\\s]*message[\\s]*=[\\s]*\"(.*?)\"[\\s]*\\)$"

func CicadaScan(relativePath string) {
	path, _ := os.Getwd()
	var paths []string
	if strings.Contains(path, "src") {
		paths = strings.Split(path, "/src")
	} else {
		paths = strings.Split(path, "/bin")
	}
	fileSet := token.NewFileSet()
	if relativePath == "" {
		relativePath = "src/exceptions"
	}
	pack, err := parser.ParseDir(fileSet, paths[0]+relativePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range pack {
		for _, file := range v.Files {
			commentMap := ast.NewCommentMap(fileSet, file, file.Comments)
			addFileError(commentMap)
		}
	}
}

func addFileError(commentMap ast.CommentMap) {
	for node, groups := range commentMap {
		for _, commentGroup := range groups {
			for _, comment := range commentGroup.List {
				if strings.Contains(comment.Text, "@CicadaError") {
					funcName := node.(*ast.Field).Names[0].Name
					r := regexp.MustCompile(annotationRegex)
					annotationInfo := r.FindStringSubmatch(comment.Text)
					code, convErr := strconv.Atoi(annotationInfo[1])
					if convErr != nil {
						log.Fatal("goCicada: Code must number")
					}
					errorHandler := func(logger *logrus.Entry, httpStatus int32, errs ...error) Throwable {
						var detail string
						for _, err := range errs {
							err.Error()
							detail = fmt.Sprintf(detail+"%s", err.Error())
						}
						throwable := Throwable{
							Status: httpStatus,
							Code:   code,
							Cause:  annotationInfo[2],
							Detail: detail,
						}
						logger.Errorln(throwable)
						return throwable
					}
					CicadaErrorMap[funcName] = errorHandler
				}
			}
		}
	}
}
