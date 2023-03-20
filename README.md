# 一个简单易用的go语言ORM框架

使用示例如下：
```go
package main

import (
    "fmt"
    "github.com/uzziahlin/orm"
)

type User struct {
    Id       int64  `orm:"id"`
    Name     string `orm:"name"`
    Password string `orm:"password"`
}

func main() {
    db := orm.Open("driverName", "dataSourceName")
    
    selecter := NewSelector[User](db)
    
    user := selecter.
		Select(C("Name"), Avg("Age")).
		From(&User{}).
		Where(C("Name").EQ("Jack")).
		Get(context.TODO())
}
```
