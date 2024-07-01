package main

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.infoLog.Println("BACKGROUND:", err)
			}
		}()
		fn()
	}()
}
