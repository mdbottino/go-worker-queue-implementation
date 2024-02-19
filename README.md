# Basic implementation of a worker queue using channels

This is a basic but functional implementation of the techniques described here: http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/

## What?

What the code does is pretty useless, it just reads the contents of a file using a worker pool (go channels) asserts the value it contains and that's it.

## Why?

If you were to read the contents of the same file sequentially it will take longer as you need to wait for the previous operation to finish before attempting to do it again. If you were to create a goroutine per iteration, performance will degrade drastically as it would hammer the disk and the OS trying to get access to the file. By using a worker pool we limit the max number of concurrent reads therefore improving the performance over sequential access and preventing degradation even if a high number of requests (to read the file) are queued

## Differences

Although it is a implementation of the techniques described in the article I've added a couple of changes to make it work for a limited (known) number of requests. In the context of a web server that just keeps receiving requests it does not need to stop when a certain number of requests have been handled. That's why I've added a `sync.WaitGroup` and a method to stop all workers once all jobs have completed. 

## Very simplistic benchmark

I won't post specifics about the equipment as this is not meant to be a "real" benchmark but rather a "ballpark figure" to give you an idea of how it improves performance:

- SSD
    * sequential: ~32 ms
    * goroutines per job: ~35-60 ms
    * worker queue: ~22 ms

- HDD
    * sequential: ~90 ms
    * goroutines per job: ~40-60 ms
    * worker queue: ~35 ms

Your mileage will certainly vary.