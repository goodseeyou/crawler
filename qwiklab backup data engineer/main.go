// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	defer cancel()

	//

	email, password := os.Getenv("QWIKLAB_EMAIL"), os.Getenv("QWIKLAB_PASSWORD")

	//for qwiklab login
	formIDValues := []FormIDValue{FormIDValue{Id: "user_email", Value: email}, FormIDValue{Id: "user_password", Value: password}}

	var res string
	err := chromedp.Run(ctx, qwiklabLogin(formIDValues, &res))
	if err != nil {
		log.Fatal(err)
	}

	for i := 63701; i <= 63724; i++ {
		goTask(ctx, i)
	}
}

func goTask(ctx context.Context, num int) {
	// capture screenshot of an element
	var buf []byte
	var strbuf string

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(ctx, fullScreenshot(fmt.Sprintf(`https://googlecloud.qwiklabs.com/classrooms/7772/labs/%d`, num), 100, &buf, &strbuf)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%d.png", num), buf, 0644); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%d.html", num), []byte(strbuf), 0644); err != nil {
		log.Fatal(err)
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Liberally copied from puppeteer's source.
//
// Note: this will override the viewport emulation settings.
func fullScreenshot(urlstr string, quality int64, res *[]byte, resStr *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		chromedp.OuterHTML(`html`, resStr),
	}
}

var urlQwiklabsSignIn = "https://googlecloud.qwiklabs.com/users/sign_in"

type FormIDValue struct {
	Id, Value string
}

func (this FormIDValue) GetSelector() string {
	return fmt.Sprintf(`#%s`, this.Id)
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
	return out
}
