# Go Channels — Mental Model

## 1. Forget the code for a moment

Imagine an unbuffered channel:

```go
ch := make(chan int)
```

There are two goroutines:

```
Goroutine A                  Goroutine B

   SEND                         RECEIVE
    │                              │
    │                              │
    └────────── rendezvous ────────┘
```

An unbuffered channel has nowhere to store the value.

Therefore:

```
Sender cannot finish
        until
Receiver is ready
```

and:

```
Receiver cannot finish
        until
Sender is ready
```

This is the fundamental behavior.

---

## 2. Scenario 1 — Receiver is already waiting

```go
package main

import "fmt"

func main() {
    ch := make(chan int)

    go func() {
        x := <-ch
        fmt.Println("received:", x)
    }()

    ch <- 42
}
```

Let's trace it.

**State A**

Main creates the channel:

```
ch = unbuffered int channel
```

**State B**

Main starts the goroutine.

Now:

```
Main
Worker
```

**State C**

Worker reaches:

```go
x := <-ch
```

Nobody has sent anything yet.

So worker blocks:

```
Worker → BLOCKED
```

The main goroutine continues.

**State D**

Main reaches:

```go
ch <- 42
```

Now the receiver is already waiting.

So the communication can happen immediately:

```
Main                         Worker

ch <- 42  ────────────────►  x := <-ch
             42
```

Worker receives 42.

The worker becomes runnable/runs and eventually prints:

```
received: 42
```

---

## 3. Scenario 2 — Sender is already waiting

Reverse the situation:

```go
package main

import "fmt"

func main() {
    ch := make(chan int)

    go func() {
        ch <- 42
    }()

    x := <-ch

    fmt.Println(x)
}
```

Worker reaches:

```go
ch <- 42
```

No receiver exists yet.

So:

```
Worker → BLOCKED
```

Then main reaches:

```go
x := <-ch
```

Now the two sides meet.

```
Worker                      Main

ch <- 42  ───────────────►  x := <-ch
             42
```

The worker's send completes.

The main's receive completes.

Main gets:

```
x = 42
```

---

## 4. Critical insight: the channel does not decide who runs

Suppose:

```go
go sender()
go receiver()
```

You might think:

> "The sender executes first because it appears first."

No.

You might think:

> "The receiver executes first because the channel needs a receiver."

Also no.

The scheduler determines which goroutine gets execution time first.

There are multiple possible paths.

**Path A**

```
sender runs
    ↓
sender blocks
    ↓
receiver runs
    ↓
communication occurs
```

**Path B**

```
receiver runs
    ↓
receiver blocks
    ↓
sender runs
    ↓
communication occurs
```

Both are valid.

The final communication relationship is the same.

---

## 5. This is the first big distinction

There are two separate concepts:

**Scheduling**

> Which goroutine runs now?

**Channel synchronization**

> Can this communication proceed now?

Don't mix them.

The scheduler decides:

```
G1 or G2 or G3?
```

The channel determines:

```
Can this send/receive proceed?
```

---

## 6. Scenario 3 — Two senders, one receiver

Now it becomes more interesting.

```go
ch := make(chan int)

go func() {
    ch <- 10
}()

go func() {
    ch <- 20
}()

x := <-ch
fmt.Println(x)
```

We have:

```
Sender A → 10
Sender B → 20
Receiver → ?
```

Suppose sender A runs first:

```go
A: ch <- 10
```

No receiver yet.

So A blocks.

Then B runs:

```go
B: ch <- 20
```

B also blocks.

Now main runs:

```go
x := <-ch
```

A receiver is finally ready.

One sender can synchronize with it.

For example:

```
A ───── 10 ─────► Receiver
```

Then:

```
A → completed
B → still blocked
```

The receiver gets 10.

The important point:

> A channel does not magically collect all the waiting senders.

One receive corresponds to one communication.

---

## 7. What about the other sender?

Sender B is still waiting.

If another receiver later appears:

```go
y := <-ch
```

B can then communicate:

```
B ───── 20 ─────► Receiver
```

This gives you an important model:

```
Each send
    ↓
needs a matching receive
```

For an unbuffered channel.

---

## 8. Scenario 4 — One sender, two receivers

Now reverse it:

```go
ch := make(chan int)

go func() {
    ch <- 99
}()

go func() {
    fmt.Println("R1:", <-ch)
}()

go func() {
    fmt.Println("R2:", <-ch)
}()
```

There is:

```
1 sender
2 receivers
```

Only one receiver can receive the single value 99.

The value is not duplicated.

It does not become:

```
R1 = 99
R2 = 99
```

Instead, exactly one receive gets the value.

Conceptually:

```
             ┌──► R1
Sender ──99──┤
             └──► R2
```

Only one path succeeds.

Which receiver gets it?

That depends on scheduling and synchronization timing. Do not assume R1 or R2.

The other receiver remains waiting for another value.

---

## 9. A channel is not broadcast

This is extremely important.

Consider:

```go
ch := make(chan string)
```

Then:

```go
ch <- "hello"
```

with three receivers.

The value is not broadcast to all three.

> A channel operation transfers a value through one send/receive communication.

If you need:

```
one event → many goroutines
```

you need a different design, often involving closing a channel, multiple channels, or another synchronization mechanism.

We'll get there later.

---

## 10. Buffered channels change the story

Now:

```go
ch := make(chan int, 2)
```

Visualize:

```
capacity = 2

┌────┬────┐
│    │    │
└────┴────┘
```

First send:

```go
ch <- 10
```

There is space.

So:

```
10
↓
┌────┬────┐
│ 10 │    │
└────┴────┘
```

The sender does not need a receiver right now.

This is the crucial difference.

---

## 11. Two sends

```go
ch <- 10
ch <- 20
```

Now:

```
┌────┬────┐
│ 10 │ 20 │
└────┴────┘
```

The buffer is full.

The sender can still have completed both sends.

No receiver was required.

---

## 12. Third send

Now:

```go
ch <- 30
```

There is no room:

```
┌────┬────┐
│ 10 │ 20 │
└────┴────┘
       FULL
```

Therefore the sender blocks.

```
Sender → BLOCKED
```

until a receiver removes something.

---

## 13. Receive from buffered channel

Suppose:

```go
x := <-ch
```

The first value is:

```
10
```

because channels behave FIFO for values waiting in the buffer.

After receiving:

```
┌────┬────┐
│ 20 │    │
└────┴────┘
```

Now there is room.

Therefore a blocked send of 30 can proceed.

Result:

```
┌────┬────┐
│ 20 │ 30 │
└────┴────┘
```

---

## 14. The exact blocking rule

Now we can formulate this precisely.

### Unbuffered channel

**Send**

```go
ch <- value
```

blocks when there is no receiver ready.

**Receive**

```go
<-ch
```

blocks when there is no sender ready.

### Buffered channel

**Send**

```go
ch <- value
```

blocks when the buffer is full.

**Receive**

```go
<-ch
```

blocks when the buffer is empty.

That's the model.

---

## 15. This table is worth memorizing

| Channel | Send | Receive |
|---|---|---|
| Unbuffered | Wait for receiver | Wait for sender |
| Buffered, space available | Proceeds | If value exists, proceeds |
| Buffered, full | Blocks | — |
| Buffered, empty | — | Blocks |

---

## 16. A practical example: producer and consumer

This is where channels become intuitive.

```go
package main

import "fmt"

func main() {
    jobs := make(chan int, 3)

    go func() {
        jobs <- 1
        jobs <- 2
        jobs <- 3
        jobs <- 4
    }()

    fmt.Println(<-jobs)
    fmt.Println(<-jobs)
    fmt.Println(<-jobs)
    fmt.Println(<-jobs)
}
```

The producer does:

```
1
2
3
4
```

The buffer has capacity 3.

So initially:

```
┌───┬───┬───┐
│ 1 │ 2 │ 3 │
└───┴───┴───┘
```

The producer then tries:

```
4
```

But the buffer is full.

Therefore:

```
Producer → BLOCKED
```

Main receives 1.

Now there is space.

So the 4 can enter.

Buffer becomes conceptually:

```
┌───┬───┬───┐
│ 2 │ 3 │ 4 │
└───┴───┴───┘
```

This leads to an important systems concept:

> A bounded channel naturally creates backpressure.

The producer cannot run infinitely ahead of the consumer.

This becomes extremely useful in networking and server design.

---

## 17. Channel capacity is a design decision

Suppose you have:

```
Producer → Channel → Consumer
```

If the channel is unbuffered:

> Producer and consumer stay tightly synchronized.

If the channel has capacity 100:

> Producer can temporarily get ahead by up to 100 queued values.

That means buffering affects:

```
latency
throughput
memory usage
backpressure
coordination
```

So this:

```go
make(chan Request)
```

versus:

```go
make(chan Request, 1000)
```

is not merely a syntax choice.

It can be an architectural decision.

---

## 18. WaitGroup vs Channel

You just learned WaitGroup.

Now compare them.

### WaitGroup

> "I'm waiting for these goroutines/work items to finish."

Example:

```go
wg.Add(3)

go worker1()
go worker2()
go worker3()

wg.Wait()
```

### Channel

> "I want goroutines to communicate and synchronize through values/events."

Example:

```go
jobs <- job
result := <-results
```

So:

```
WaitGroup → completion coordination

Channel   → communication + synchronization
```

They are not interchangeable.

And in real programs, they are often used together.

---

## 19. Example: networking server

Imagine a TCP server.

You might have:

```
             ┌──────────────┐
Client 1 ───►│              │
Client 2 ───►│ TCP Server   │
Client 3 ───►│              │
Client 4 ───►│              │
             └──────┬───────┘
                    │
                    ▼
                 jobs
                 channel
                    │
          ┌─────────┼─────────┐
          ▼         ▼         ▼
        Worker 1  Worker 2  Worker 3
```

The accepting goroutine can send work:

```go
jobs <- conn
```

Workers receive:

```go
conn := <-jobs
```

Now channels become a central piece of real concurrent architecture.

We'll eventually build this pattern ourselves.

---

## 20. A powerful debugging technique

When you see confusing channel code, don't ask:

> "What does the program probably do?"

Instead, make a table.

For example:

```go
ch := make(chan int, 2)

go func() {
    ch <- 10
    ch <- 20
    ch <- 30
}()

fmt.Println(<-ch)
fmt.Println(<-ch)
fmt.Println(<-ch)
```

Trace it:

| Event | Buffer | Producer |
|---|---|---|
| Start | empty | running |
| send 10 | [10] | continues |
| send 20 | [10,20] | continues |
| send 30 | full | blocks |
| receive | [20] | wakes/progresses |
| receive | [30] | — |
| receive | [] | — |

This is how you should mentally execute channel programs.

---

## 21. A subtle but important point

A blocked goroutine is not necessarily consuming CPU.

Suppose:

```go
x := <-ch
```

and no value exists.

That goroutine waits.

It does not sit there continuously executing:

```
check
check
check
check
check
```

like a naive busy loop.

The runtime can suspend that goroutine and execute other runnable goroutines.

This is one reason goroutine-based concurrency works efficiently.

---

## 22. Another subtle point: blocking is local

Suppose:

```
G1 → blocked receiving from ch
G2 → running
G3 → running
G4 → waiting on network I/O
```

A blocked G1 does not mean:

```
ENTIRE PROGRAM BLOCKED
```

It means:

```
G1 cannot currently make progress.
```

The runtime can continue scheduling other work.

This distinction is exactly the same mental discipline we used with goroutine states.

---

## 23. The complete mental model so far

You can now think of channels like this:

```
                    CHANNEL
                       │
          ┌────────────┴────────────┐
          │                         │
      COMMUNICATION            SYNCHRONIZATION
          │                         │
       values                  blocking/wakeup
          │                         │
          └────────────┬────────────┘
                       │
                  goroutines
```

And for blocking:

```
UNBUFFERED

send ───── waits for ───── receive
receive ── waits for ───── send


BUFFERED

send ─────► buffer ─────► receive
              │
          full → sender blocks
          empty → receiver blocks
```

---

## Checkpoint 2 — Trace, don't guess

Don't run these. Reason about them.

### Q1

```go
ch := make(chan int)

go func() {
    ch <- 10
    fmt.Println("A")
}()

fmt.Println("B")
```

Does "A" necessarily print?

Why or why not?

### Q2

```go
ch := make(chan int)

go func() {
    fmt.Println("A")
    ch <- 10
    fmt.Println("B")
}()

x := <-ch
fmt.Println("C", x)
```

What ordering relationship is guaranteed among:

```
A
B
C
```

Be precise.

### Q3

```go
ch := make(chan int, 2)

ch <- 1
ch <- 2
ch <- 3

fmt.Println("done")
```

Exactly where does the main goroutine block?

### Q4

Consider:

```go
ch := make(chan int)

go func() {
    ch <- 10
}()

go func() {
    ch <- 20
}()

x := <-ch
fmt.Println(x)
```

Can x be 10?

Can x be 20?

Can it be both?

### Q5

One sender sends:

```go
ch <- 100
```

There are two receivers waiting.

Does the value go to both receivers, one receiver, or neither?

Explain.

### Q6

Why does this work without an immediately available receiver?

```go
ch := make(chan int, 5)

ch <- 10
```

What is fundamentally different from:

```go
ch := make(chan int)

ch <- 10
```

### Q7 — Architecture question

Suppose you have:

```
Producer → Channel → Consumer
```

What happens to the producer when:

```
unbuffered channel
```

is used?

And what changes when:

```
buffered channel with capacity 100
```

is used?

Explain it in terms of synchronization and backpressure, not merely storage.

### Q8 — Most important

Complete this mental rule:

> For an unbuffered channel, a send and receive ________ each other.
>
> For a buffered channel, the buffer can temporarily ________ the sender and receiver.