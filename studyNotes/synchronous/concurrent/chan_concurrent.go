package chan_concurrent

func Concurrent(MaxCG int64) {
    ch := make(chan struct{}, MaxCG)
    
    for i := 0; i < MaxCG; i++ {
        ch <- struct{}{}
    }

    done := make(chan bool)
    waitAllJobs := make(chan bool)
    go func(){
        for {
            <-done

            ch <- struct{}{}
        }
        waitAllJobs <- true
    }()

    for {
        <-ch 

        go func() {
            // Todo something
            
            done <- true
        }()
    }

    <-waitAllJobs
}
