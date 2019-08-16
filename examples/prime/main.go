package main

import (
	"context"
	"log"
	"sync"

	"github.com/discoverkl/lorca"
)

const html = `data:text/html,
<html>
<title>Primes</title>
<div id="container" class="root">
	<div class="head">How many primes do you want? &nbsp; &nbsp; <input autofocus onkeyup="run(this)"/></div>
	<pre id="output"></pre>
</div>
<script type="text/javascript">
let no = 1
function println(msg, lineno) {
	let line = (lineno ? (no++).toString().padStart(5, "0") + ": " : "")
	document.getElementById("output").innerText += (line + msg + "\n")
	let container = document.getElementById("container")
	container.scrollTo(container.scrollLeft, container.scrollHeight);
}
function clear() {
	no = 1
	document.getElementById("output").innerText = ""
}

let lastTimer;
let last = 0;
function run(e) {
	if (lastTimer) clearTimeout(lastTimer)
	lastTimer = setTimeout(() => {
		let count = (e.value.length == 0 ? 0 : parseInt(e.value))
		if (!(count >= 0) || count == last) return
		last = count
		clear()
		lineno = (count > 10)
		js2go(count, (a) => {
			println(a.join(" "), lineno)
		})
	}, 10)
}
</script>

<style type="text/css">
html, body {
	height: 100%;
	margin: 0;
	overflow: hidden;
}
.root {
	height: 100%;
	overflow: scroll;
}
.head {
	position: absolute;
	background: white;
	width: 600px;
	height: 20px;
	margin-right: 50px;
	padding: 0.5em;
}
pre {
	margin-top: 40px;
	padding-left: 0.5em;
}
input {
	border-width: 0;
    border-bottom-width: 1px;
    outline: none;
    text-align: center;
    width: 60px;
    border-color: black;
}
</style>
</html>
`

var cancelActiveJob context.CancelFunc
var jobLock sync.Mutex

func main() {
	ui, err := lorca.New("", "", 800, 600, "--window-position=200,200")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	ui.Bind("js2go", js2go)
	ui.Load(html)
	<-ui.Done()
}

func js2go(count int, fn *lorca.Function) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobLock.Lock()
	if cancelActiveJob != nil {
		cancelActiveJob()
	}
	cancelActiveJob = cancel
	jobLock.Unlock()

	ch := Prime(count)
	buffer := []int{}
	i := 0
loop:
	for {
		select {
		case prime, ok := <-ch:
			if !ok {
				break loop
			}
			i++
			buffer = append(buffer, prime)
			if len(buffer) >= 10 {
				fn.Call(buffer)
				buffer = buffer[0:0]
			}
		case <-ctx.Done():
			return
		}
	}
	if len(buffer) > 0 {
		fn.Call(buffer)
	}
}

// A concurrent prime sieve

// Send the sequence 2, 3, 4, ... to channel 'ch'.
func Generate(ch chan<- int) {
	for i := 2; ; i++ {
		ch <- i // Send 'i' to channel 'ch'.
	}
}

// Copy the values from channel 'in' to channel 'out',
// removing those divisible by 'prime'.
func Filter(in <-chan int, out chan<- int, prime int) {
	for {
		i := <-in // Receive value from 'in'.
		if i%prime != 0 {
			out <- i // Send 'i' to 'out'.
		}
	}
}

// The prime sieve: Daisy-chain Filter processes.
func Prime(count int) chan int {
	ret := make(chan int)
	ch := make(chan int) // Create a new channel.
	go Generate(ch)      // Launch Generate goroutine.
	go func() {
		defer close(ret)
		for i := 0; i < count; i++ {
			prime := <-ch
			ret <- prime
			ch1 := make(chan int)
			go Filter(ch, ch1, prime)
			ch = ch1
		}
	}()
	return ret
}
