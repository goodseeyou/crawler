package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

var urlQwiklabsSignIn = "https://googlecloud.qwiklabs.com/users/sign_in"

type FormIDValue struct {
	Id, Value string
}

func (this FormIDValue) GetSelector() string {
	return fmt.Sprintf(`#%s`, this.Id)
}

func main() {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	defer cancel()

	email, password := os.Getenv("QWIKLAB_EMAIL"), os.Getenv("QWIKLAB_PASSWORD")

	//for qwiklab login
	formIDValues := []FormIDValue{FormIDValue{Id: "user_email", Value: email}, FormIDValue{Id: "user_password", Value: password}}

	var res string
	err := chromedp.Run(ctx, qwiklabLogin(formIDValues, &res))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(strings.TrimSpace(res))
}

func genWaitVisibleQueryActions(formIDValues []FormIDValue) []chromedp.QueryAction {
	out := make([]chromedp.QueryAction, 0, len(formIDValues))
	for _, formIDValue := range formIDValues {
		out = append(out, chromedp.WaitVisible(formIDValue.GetSelector(), chromedp.ByID))
	}
	return out
}

func genSendKeysQueryActions(formIDValues []FormIDValue) []chromedp.QueryAction {
	out := make([]chromedp.QueryAction, 0, len(formIDValues))
	for _, formIDValue := range formIDValues {
		out = append(out, chromedp.SendKeys(formIDValue.GetSelector(), formIDValue.Value, chromedp.ByID))
	}
	return out
}

func qwiklabLogin(credentials []FormIDValue, res *string) chromedp.Tasks {
	out := make(chromedp.Tasks, 0)

	out = append(out, chromedp.Navigate(urlQwiklabsSignIn))

	for _, queryAction := range genWaitVisibleQueryActions(credentials) {
		out = append(out, queryAction)
	}

	for _, queryAction := range genSendKeysQueryActions(credentials) {
		out = append(out, queryAction)
	}

	//out = append(out, chromedp.Click(`//button[@type="submit"]`))
	out = append(out, chromedp.Submit(credentials[0].GetSelector(), chromedp.ByID))

	//out = append(out, chromedp.WaitNotVisible(credentials[0].GetSelector(), chromedp.ByID))
	//out = append(out, chromedp.WaitVisible(`.alert alert-error flash`), chromedp.b)
	out = append(out, chromedp.Sleep(time.Second*10))
	out = append(out, chromedp.Navigate("https://googlecloud.qwiklabs.com/classrooms/7772/labs/63701"))
	out = append(out, chromedp.Sleep(time.Second*30))
	out = append(out, chromedp.OuterHTML(`html`, res))
	return out
}
