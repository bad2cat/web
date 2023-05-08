# SSE
正常情况下，HTTP 服务器的响应只能是客户端发出请求的那个 HTML 文档。一旦发送完毕，连接就会关闭，服务器也就不会再给客户端发送任何信息了。但是，现在有一种技术，允许服务器不断向客户端发送信息，这种技术就叫做服务器推送（Server Push）。

这种技术能够将返回的数据作为一个数据流，不断地返回给客户端。这种技术的好处是，客户端不需要询问服务器，就能获得新的数据。这种技术的原理是，服务器与客户端之间建立一种连接，两者就像两个人之间的电话一样，一端打开嘴说话，另一端就能听到声音。服务器就像电话那端的人，客户端就像听筒这端的人。只要服务器不挂断电话，客户端就能随时听到服务器传过来的信息。

SSE 就是利用这种机制，基于`HTTP`协议，使用流信息向服务器发送数据。它的用法非常简单，就是在`HTTP`协议之上，使用`EventSource`对象，服务器向客户端发送数据，客户端监听`message`事件就行了。

## SSE 与 WebSocket 的比较
SSE 与 WebSocket 都是服务器向客户端推送数据的方法，所以看起来很相似。它们的不同之处在于：

- SSE 使用 HTTP 协议，而 WebSocket 使用自定义协议。
- SSE 是单向通道，只能服务器向客户端发送数据，反过来不行；WebSocket 是双向通道，双方可以随时发送数据。
- SSE 的连接断开后，浏览器会自动发起重连；WebSocket 的连接一旦断开，就会一直断开。
- SSE 默认支持断线重连，WebSocket 需要自己实现。
- SSE 一般用来传送文本，二进制数据需要编码后发送，WebSocket 默认支持二进制数据。
- SSE 一般只用来传送少量数据，因为频繁的传送会对服务器造成压力；WebSocket 传送的数据量没有限制。

## SSE 的用法

### 服务器端

服务器端的代码非常简单，就是发送一个`HTTP`头信息，然后发送数据。下面是一个例子。

```go
package main

import (
    "fmt"
    "net/http"
    "time"
)

func main() {
    http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
        flusher, ok := w.(http.Flusher)
        if !ok {
            http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        w.Header().Set("Access-Control-Allow-Origin", "*")

        for i := 0; i < 10; i++ {
            fmt.Fprintf(w, "data: Message %d\n\n", i)
            flusher.Flush()
            time.Sleep(1 * time.Second)
        }
    })

    http.ListenAndServe(":8080", nil)
}
```
上面代码中，`http.ResponseWriter`接口有一个`Flush`方法，可以将缓冲区的数据发送到客户端。这是 SSE 的关键，也是 SSE 与普通`HTTP`连接的不同之处。

### 客户端

客户端的代码也很简单，就是监听`message`事件。

```js

var source = new EventSource('http://localhost:8080/sse');

source.onmessage = function(e) {
    console.log(e.data);
};
```
上面代码中，`EventSource`对象的`onmessage`属性指定监听`message`事件的回调函数。一旦服务器发来新的数据，这个函数就会执行。

## SSE 数据格式

服务器向浏览器发送的 SSE 数据， 必须是 UTF-8 编码的文本。它由一系列字段组成，每个字段占一行。字段之间，以一个空行分隔，每一行的格式如下:

```http
field: value\n
```
上面的`field`可以取的值如下:

- `event`: 事件的名称，必须是字符串，如果没有该字段，默认为`message`事件。该字段对应的值，会被放在`Event`属性上面，例如：`event: userconnect\n\n`。
- `data`: 事件的数据，可以是字符串，也可以是二进制数据（比如`Blob`对象）。该字段对应的值，会被放在`data`属性上面，例如：`data: {"username": "bobby", "time": "02:33:48"}\n\n`。
- `id`: 消息的 ID，必须是字符串，例如：`id: 12345\n`。该字段对应的值，会被放在`lastEventId`属性上面。
- `retry`: 两条消息之间的时间间隔，单位是毫秒。浏览器接收到一条消息后，会等待指定的毫秒数，再向服务器发出下一条消息，例如：`retry: 10000\n`。
- `comment`: 注释行，必须在冒号后面有一个空格,例如：`: test\n`。

下面是一个例子：

```http
event: userconnect\n\n
data: some text\n\n
id: 12345\n
retry: 10000\n
```

发送的请求头，如下: 

```http
HTTP/1.1 200 OK

Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Origin: *
```
在上面的`Header`中，必须要指定`Content-Type: text/event-stream`

## SSE 状态管理

SSE 有一个`readyState`属性，表示连接的状态，它有三个值：

- `EventSource.CONNECTING`：连接尚未建立，或者连接已经断开，正在重连。
- `EventSource.OPEN`：连接已经建立，可以传输数据。
- `EventSource.CLOSED`：连接已经关闭，不可传输数据。

一般是在`onopen`和`onerror`事件的回调函数中，使用这个属性。

```js

var source = new EventSource('http://localhost:8080/sse');

source.onopen = function(e) {
    console.log('Connection was opened.');
};

source.onerror = function(e) {
    if (this.readyState == EventSource.CONNECTING) {
        console.log('Connection was lost, trying to reconnect...');
    } else {
        console.log('An error occurred, now closing the connection.');
        this.close();
    }
};
```

在服务器端关闭`SSE`连接，只要发送一个`HTTP`头信息即可。

```http

HTTP/1.1 200 OK

Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Origin: *
Connection: close
```

在客户端关闭`SSE`连接， 只要调用`EventSource`对象的`close`方法即可。

```js

var source = new EventSource('http://localhost:8080/sse');
// close connection
source.close();
```
