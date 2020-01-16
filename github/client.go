/*
@Time 2019-07-26 09:52
@Author ZH

*/
package github

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/zhcppy/github-console/console"
	"github.com/zhcppy/gojson"
	"golang.org/x/oauth2"
)

var clientMap = make(map[string]*github.Client)
var mu sync.Mutex

type User struct {
	token string
}

func NewUser(token string) (u *User, err error) {
	if token == "" {
		//token, err := terminal.ReadPassword(syscall.Stdin)
		token, err = console.Stdin.PromptPassword("Your github token: ")
		if err != nil {
			return nil, err
		}
	}
	if len(token) != 40 {
		return nil, errors.New("invalid github token: " + token)
	}
	return &User{token: token}, nil
}

func (u *User) NewGithubClient(ctx context.Context) *github.Client {
	mu.Lock()
	defer mu.Unlock()
	if client, ok := clientMap[u.token]; ok && client != nil {
		return client
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: u.token})
	c := github.NewClient(oauth2.NewClient(ctx, ts))
	clientMap[u.token] = c
	return c
}

func (u *User) ExecCommand(ctx context.Context, input string) error {
	hubClient := u.NewGithubClient(ctx)
	var (
		field  string
		method string
		params = []reflect.Value{reflect.ValueOf(ctx)}
		t      int
	)

	for i, k := range input {
		if k == '.' {
			field = input[t:i]
			t = i
			continue
		}
		if k == '(' {
			method = input[t+1 : i]
			t = i
			continue
		}
		if k == ',' || k == ')' {
			params = append(params, parseParams(input[t+1:i]))
			t = i
		}
	}
	if field == "" || method == "" {
		return errors.New("error: Not Fetch Service or Method")
	}

	fmt.Println("field:", field, "| method:", method, "| params:", params)
	results := reflect.ValueOf(hubClient).Elem().FieldByName(field).MethodByName(method).Call(params)
	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	for i := 0; i < len(results)-1; i++ {
		if res := results[i]; checkResult(res) {
			data := gojson.MustMarshal(res.Interface())
			fmt.Printf("%02d - %s:\n%s\n", i+1, reflect.Indirect(res).Type(), data)
		}
	}
	return nil
}

func checkResult(res reflect.Value) bool {
	return res.IsValid() && !res.IsNil() && ignoreResponse(res)
}

func ignoreResponse(res reflect.Value) bool {
	return reflect.Indirect(res).Type().String() != "github.Response"
}

func parseParams(params string) reflect.Value {
	switch params {
	default:
		return reflect.ValueOf(params)
	}
}

func WordCompleter() (word []string) {

	var data = make(map[string]map[string][]string)

	clientType := reflect.TypeOf(github.Client{})
	if clientType.Kind() == reflect.Ptr {
		clientType = clientType.Elem()
	}

	for i := 0; i < clientType.NumField(); i++ {
		field := clientType.Field(i)
		if !strings.HasSuffix(field.Type.String(), "Service") {
			continue
		}
		data[field.Name] = make(map[string][]string)
		word = append(word, field.Name)

		methodTp := reflect.New(field.Type).Type()
		if methodTp.Kind() == reflect.Ptr {
			methodTp = methodTp.Elem()
		}

		for j := 0; j < methodTp.NumMethod(); j++ {
			if methodTp.Method(j).Type.Kind() != reflect.Func {
				continue
			}
			word = append(word, fmt.Sprintf("%s.%s", field.Name, methodTp.Method(j).Name))

			params := ""
			for k := 2; k < methodTp.Method(j).Type.NumIn(); k++ {
				data[field.Name][methodTp.Method(j).Name] = append(data[field.Name][methodTp.Method(j).Name], methodTp.Method(j).Type.In(k).Name())
				params += methodTp.Method(j).Type.In(k).String()
				if k+1 != methodTp.Method(j).Type.NumIn() {
					params += ", "
				}
			}
			word = append(word, fmt.Sprintf("%s.%s(%s)", field.Name, methodTp.Method(j).Name, params))

			data[field.Name][methodTp.Method(j).Name] = append(data[field.Name][methodTp.Method(j).Name], "")

			for k := 0; k < methodTp.Method(j).Type.NumOut(); k++ {
				data[field.Name][methodTp.Method(j).Name] = append(data[field.Name][methodTp.Method(j).Name], methodTp.Method(j).Type.Out(k).Name())
			}
			data[field.Name][methodTp.Method(j).Name] = append(data[field.Name][methodTp.Method(j).Name], methodTp.Method(j).Type.String())
		}
	}
	return
}
