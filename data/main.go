// Command data takes a body of text and measures the
// characters therein to create training data.
// It uses a remote instance of Google Chrome to measure
// the text, since browsers have various text-measuring
// abilities built-in.
//
// To use this, you will want to run Chrome with the flag:
//
//     --remote-debugging-port=9222
//
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/muniverse/chrome"
)

type Sample struct {
	Input  string
	Widths string
}

func main() {
	var chromeHost string
	var dataFile string

	flag.StringVar(&chromeHost, "chrome", "localhost:9222", "Chrome DevTools host")
	flag.StringVar(&dataFile, "data", "val_seed.txt", "text to seed data")
	flag.Parse()

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		essentials.Die(err)
	}

	ep := findPage(chromeHost)
	conn, err := chrome.NewConn(context.Background(), ep.WebSocketURL)
	if err != nil {
		essentials.Die(err)
	}
	defer conn.Close()

	if err != setupCanvas(conn) {
		essentials.Die(err)
	}

	samples := make(chan *Sample, runtime.GOMAXPROCS(0))
	inputs := splitData(data)
	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			genMeasurements(conn, inputs, samples)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(samples)
	}()
	for sample := range samples {
		// TODO: print the sample here.
		fmt.Println(sample.Input)
		fmt.Println(sample.Widths)
	}
}

func findPage(host string) *chrome.Endpoint {
	endpoints, err := chrome.Endpoints(context.Background(), host)
	if err != nil {
		essentials.Die(err)
	}
	for _, ep := range endpoints {
		if ep.Type == "page" {
			return ep
		}
	}
	essentials.Die("no pages open")
	return nil
}

func setupCanvas(conn *chrome.Conn) error {
	err := conn.EvalPromise(context.Background(), "Promise.resolve("+
		"(window.mtcontext = document.createElement('canvas')"+
		".getContext('2d'), window.mtcontext.font = '12px Arial', true));", nil)
	return essentials.AddCtx("setup canvas", err)
}

func splitData(data []byte) <-chan []byte {
	res := make(chan []byte, 1)
	go func() {
		defer close(res)
		var buf bytes.Buffer
		for _, b := range data {
			if b != '\n' {
				buf.WriteByte(b)
			}
			if buf.Len() > 0 && rand.Intn(10) == 0 {
				res <- append([]byte{}, buf.Bytes()...)
				buf.Reset()
			}
		}
	}()
	return res
}

func genMeasurements(c *chrome.Conn, in <-chan []byte, out chan<- *Sample) {
	for sampleIn := range in {
		var ints []int
		for _, ch := range sampleIn {
			ints = append(ints, int(ch))
		}
		charJSON, _ := json.Marshal(ints)
		code := `Promise.resolve((function(chars) {
			var ctx = window.mtcontext;
			var measureForward = function(chars) {
				var str = '';
				var lastWidth = 0;
				var res = [];
				return chars.map((ch) => {
					str += String.fromCharCode(ch);
					var w = ctx.measureText(str).width;
					var res = w - lastWidth;
					lastWidth = w;
					return res;
				});
				return res;
			};
			var forw = measureForward(chars);
			var back = measureForward(chars.reverse());
			return forw.map((x, i) => (x + back[i]) / 2);
		})(` + string(charJSON) + `));`
		var widths []float64
		err := c.EvalPromise(context.Background(), code, &widths)
		if err != nil {
			essentials.Die("measure text:", err)
		}
		out <- &Sample{
			Input:  oneHotStr(sampleIn),
			Widths: numbersStr(widths),
		}
	}
}

func oneHotStr(data []byte) string {
	var res []string
	for _, b := range data {
		for i := 0; i < 0x100; i++ {
			if i == int(b) {
				res = append(res, "1")
			} else {
				res = append(res, "0")
			}
		}
	}
	return strings.Join(res, " ")
}

func numbersStr(nums []float64) string {
	var res []string
	for _, n := range nums {
		res = append(res, strconv.FormatFloat(n, 'f', 5, 64))
	}
	return strings.Join(res, " ")
}
