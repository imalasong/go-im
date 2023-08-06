package main

func main() {

	sever := NewServer("127.0.0.1", 8888)

	sever.Start()

}
