package closure


func Closure(x int) func() {
    println("test.x:", &x)

    return func() {
        println("closure.x :", &x, x)
    }
}
