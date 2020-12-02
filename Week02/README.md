# Errors in Go

## Errors vs Exception
### Error
* Go error 是一个普通的接口，而不是语言提供的机制
* errors.New() 返回的是内部 errorString 对象的指针，防止两个不同的对象在比较时出现相等的情况
### Why not Exception
* 各语言演进史
    * c语言：单返回值，一般指针作为入参，返回值int表示成功还是失败
    * c++：引入了 exception，但是无法知道被调用方会抛出什么异常
    * java：引入了 checked exception，方法的所有者必须声明，调用者必须处理。
    带来的问题是：可能会被程序员滥用，比如使用根 Exception 捕获却不处理；异常导致程序处理流程发生跳转。
* Go的错误处理思想
    * 不引入 exception，而是支持多参返回，一般通过最后一个返回 error
    * 如果一个函数返回 error，必须先对 error 进行判断。除非对返回的 value 不关心，否则在 error 判定失败的情况下不能对 value 做任何假设
    * panic 意味着程序不能继续执行，所以不能假设调用者会解决 panic。通常只有在索引越界，不可恢复环境问题，栈溢出时才使用 panic。
* Go错误处理的好处
    * 简单
    * 考虑失败，而不是成功
    * 没有隐藏的控制流
    * 使用者控制error
    * errors are values
###  Practice
1. 使用 bool 值作为异常结果(错误实践)
    ```go
    func Positive(n int) (bool, bool) {
        if n == 0 {
            return false, false
        }
        return n > -1, true
    }

    func Check(n int) {
        pos, ok := Positive(n)
        if !ok {
            fmt.Println(n, "is neither")
            return
        }
        if pos {
            fmt.Println(n, "is positive")
        } else {
            fmt.Println(n, "is negative") 
        }
    }
    ```
2. 使用

## Error Type
### Sentinel Error
预定义的错误称为Sentinel Error。这是最不灵活的错误处理策略，因为调用方必须使用 == 将结果与预先声明的值进行比较。
当需要提供更多的上下文时，返回一个不同的错误将破坏相等性检查，迫使调用者查看error.Error()方法的输出，以查看它是否
匹配特定的字符串。

实践指导：
* 程序错误处理不应该依赖error.Error()的输出，它的输出用于记录日志，输出到stdout等。
* 如果公共函数返回一个特定值的错误，那么该值必须是公共的，并且需要文档记录；如果API定义返回了一个特定错误的
interface，那么接口的所有实现都将被限制为仅返回该错误，即使它们可以提供更具描述性的错误。
* 如果许多包都导出错误值，那么项目中容易出现其他包必须导入这些错误值才能检查特定类型的错误，导致包的依赖变复杂。
* 尽可能避免 sentinel error，尽管标准库中有一些使用了它们的情况。
### User Error Types
Error type是实现了 error 接口的自定义类型，它们能够包装底层错误，提供更多的上下文。它比 sentinel error 要好一些，
但是它们仍然未解决 sentinal error 的一些问题：调用者需要使用类型断言和类型 switch，导致调用者与被调用者产生强耦合，
从而使API变得脆弱。所以尽量避免错误类型，或者至少避免将它们作为公共API的一部分。
### Opaque Error
不透明的错误提供的是：只需返回错误而不假设其内容。也就是调用者只知道关于操作结果是成功还是失败，但是看不到错误的内部。

在少数情况下，这种二分错误处理的方式是不够的。例如，与进程外的世界进行交互，需要调用方调查错误的性质，以确定重试该操作
是否合理。这种情况下，我们可以断言错误实现了特定的行为，而不是断言错误是特定的类型或值。
```go
type temporary interface {
    Temporary() bool
}

func IsTemporary(err error) bool {
    te, ok := err.(temporary)
    return ok && te.Temporary()
}
```

## Handling Error
### Indented flow is for errors
无错误的正常流程代码是一条直线，而不是缩进的代码。
```go
f, err := os.Open(path)
if err != nil {
    // handle error
}
// do stuff

f, err := os.Open(path)
if err == nil {
  // do stuff
}
// handle error
```
### Eliminate error handling by eliminating errors
移除不必要的代码：
```go
// bad
func AuthenticateRequest(r *Request) error {
    err := authenicate(r.User)
    if err != nil {
        return err
    }
    return nil
}

// good
func AuthenticateRequest(r *Request) error {
    return authenticate(r.User)
}
```
使用封装程度更好的库函数：
```go
// bad
func CountLines(r io.Reader) (int, error) {
    var (
        br = bufio.NewReader(r)
        lines int
        err error
    )
    
    for {
        _, err = br.ReadString('\n')
        lines++
        if err != nil {
            break
        }
    }
 
    if err != io.EOF {
        return 0, err
    }
    return lines, nil
}

// good
func CountLines(r io.Reader) (int, error) {
    sc := bufio.NewScanner(r)
    lines := 0
    
    for sc.Scan() {
        lines++
    }

    return lines, sc.Err()
}
```
封装错误：
```go
// bad
type Header struct {
    Key, Vlaue string
}

type Status struct {
    Code int
    Reason string
}

func WriteResponse(w io.Writer, st Status, headers []Header, body io.Reader) error {
    _, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", st.Code, st.Reason)
    if err != nil {
        return err
    }
  
    for _, h := range headers {
        _, err := fmt.Fprintf(w, "%s: %s\r\n", h.Key, h.Value)
        if err != nil {
            return err
        }
    }

    if _, err := fmt.Fprint(w, "\r\n"); err != nil {
        return err
    }

    _, err = io.Copy(w, body)
    return err
}

// good
type errWriter struct {
    io.Writer
    err error
}

func (e *errWriter) Write(buf []byte) (int, error) {
    if e.err != nil {
        return 0, e.err    
    }

    var n int
    n, e.err = e.Writer.Write(buf)
    return n, nil
}

func WriteResponse(w io.Writer, st Status, headers []Header, body io.Reader) error {
    ew := &errWriter{Writer: w}
    fmt.Fprintf(ew, "HTTP/1.1 %d %s\r\n", st.Code, st.Reason)
   
    for _, h := range headers {
        fmt.Fprintf(ew, "%s: %s\r\n", h.Key, h.Value)
    }

    fmt.Fprint(ew, "\r\n")
    io.Copy(ew, body)

    return ew.err
}
```
### Wrap errors
像之前authenticate那样直接将错误返回给调用方，调用者也可能这样做，最后导致程序的顶部得到的只是最底层的错误信息。
因为没有调用堆栈的跟踪信息，所以代码的作者将被迫进行长时间的代码分割，以发现是哪个代码路径触发了最底层的错误。
```go
func AuthenticateRequest(r *Request) error {
    err := authenticate(r.Uesr)
    if err != nil {
        return fmt.Errorf("authenticate failed: %v", err)
    }
    return nil
}
```
正如前面介绍的，这种模式与sentinel errors或type assertions不兼容，会导致错误等值判定失败。

我们经常发现类似的代码，在错误处理中，带了两个任务：记录日志并且再次返回错误。上层代码遇到错误
之后再次加上当前的函数的信息将错误记录到日志，然后再次将错误返回。比如：
```go
func WriteAll(w io.Writer, buf []byte) error {
    _, err := w.Write(buf)
    if err != nil {
        log.Println("unable to write:", err)
        return err
    }
    return nil
}

func WriteConfig(w io.Writer, conf *Config) error {
    buf, err := json.Marshal(conf)
    if err != nil {
        log.Printf("could not marshal config: %v", err)
        return err
    }
    if err := WriteAll(w, buf); err != nil {
        log.Println("could not write config: %v", err)
        return err
    }
    return nil
}

func main() {
    err := WriteConfig(f, &conf)
    fmt.Println(err)
}
```
这段代码造成的后果是最后同一个错误在不同的函数栈帧上输出到日志，而最顶层只得到一个最底层的错误。
针对错误处理，要遵循的一个原则是：*错误应该只被处理一次，并且在遇到错误时不能对函数返回值做出任何假设*。

在实践时，日志记录与错误无关且对调试没有帮助的信息应被是为噪音，应予以质疑。日志只记录失败的原因。
对于错误处理，需要做到：
* 错误要被日志记录
* 应用程序处理错误，保证100%完整性
* 之后不再报告当前错误

### 使用pkg/errors库
pkg/errors能够将堆栈保存起来，也可以加上上下文。它主要有两个函数Wrap和WithMessage，Wrap可以加上
堆栈信息和上下文，WithMessage可以加上上下文。该库使用时容易出现误用，需要遵循以下原则：
* 在应用代码中使用errors.New或者errors.Errorf返回错误
* 如果调用项目中其他包内的函数，通常直接返回
* 如果和其他库进行协作，考虑使用errors.Wrap或者errors.Wrapf保存堆栈信息。同样适用于和标准库协作
* 直接返回错误，而不是每个错误产生的地方到处打日志
* 在程序的顶部或者是工作的goroutine顶部(请求入口)，使用%+v把堆栈详情记录
* 使用errors.Cause获取root error，再进行和sentinel error判定
总结：
* 选择wrap error是只有应用可以选择应用的策略，而具有最高可重用性的包只能返回根错误值
* 如果函数/方法不打算处理错误，那么用足够的上下文wrap errors并将其返回到调用堆栈中。上下文可以包括
使用的参数或者失败的查询语句，确定记录的上下文是足够多还是太多的一个方法是检查日志并验证它们在开发期间
是否为您工作
* 一旦确定函数/方法将处理错误，错误就不再是错误。如果函数/方法仍然需要发出返回，则它不能返回错误值。
它应该只返回零(比如降级处理中，返回了降级数据，需要返回nil)。

## Go 1.13 errors
Go 1.13为errors和fmt引入了新特性，以简化处理包含其他错误的错误。其中最重要的是：包含另一个错误的error
可以实现返回底层错误的Unwrap方法。如果e1.Unwrap()返回e2，那么我们说e1包装e2，可以展开e1获取e2。

Go 1.13 errors包包含两个用于检查错误的新函数：Is和As。
```go
// Similar to: if err == ErrNotFound {...}
// Is will call err.Unwrap if err implements Unwrap
if errors.Is(err, ErrNotFound) {...} 

// Similar to: if e, ok := err.(*QueryError); ok {...}
var e *QueryError
if errors.As(err, &e) {...}
```

Go 1.13支持fmt.Errorf支持新的%w谓词，它可以用于errors.Is和errors.As。
```go
if err != nil {
    // Return an error which unwraps to err
    reutrn fmt.Errorf("decompress %v: %w", name, err)
}

err := fmt.Errorf("access denied: %w", ErrPermission)
...
if errors.Is(err, ErrPermission) {...}
```

新特性实践：
* 自定义的错误可以提供Is和As方法，以方便使用者比较
* 包的API使用fmt.Errorf的%w将sentinal error包装之后返回，使得API可以在之后修改错误信息
* Go 1.13未支持堆栈信息，所以pkg/errors对Go 1.13做了兼容

## Go 2 Error Inspection
// TODO: 补充说明
https://go.googlesource.com/proposal/+/master/design/29934-error-values.md