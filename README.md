<img alt="gus" src="http://imgur.com/UUpaKM4.jpg" height="220" />

gus  
========

A simple user authentication and account management library with go/sql compatible implementations.

Benefits
========

* No http or UI dependencies
* Ideal for lightweight applications
* Can use default sqlite3 driver without setting up a database
* Low resource requirements
* Easily compose in an existing application or extend to add additional functionality


Features
========

* Local user authentication
    * Sign-up, Sign-in
    * Change and reset password
    * PLANNED: Locking with Rate limit locking
* User management
* Basic Organisation management

Get started
========
```bash
go get github.com/kwk-super-snippets/gus 
```

```go
 // Min config
 o := gus.DbOptions{DataSourceName: "./my.db", Seed: true}
 db = gus.GetDb(o)
 users := gus.NewUsers(db)
	
 // Create user
 p := CreateUserParams{Email:"some@email.com"}
 user, tempPass, err := users.Create(p)
	
 // Sign-in
 p := &gus.SignInParams{Email:user.Email, Password:tempPass}
 u, err := users.Authenticate(*p)
```    

Logging
--
By default debug logging is enabled you can either provide your own implementation *log.Logger
```go
gus.DebugLogger = log.New(os.Stdout, "GUS: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
```
Or set it to nil to disable logging
```go
gus.DebugLogger = nil
```
The same goes for `gus.ErrorLogger`