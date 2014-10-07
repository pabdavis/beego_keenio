Keen IO Middleware for Beego Framework
======================================

Keen IO Middleware for Beego Framework. The [Keen IO](https://keen.io/) API lets developers build analytics features directly into their apps.

[ ![Codeship Status for pabdavis/beego_keenio](https://codeship.io/projects/88b34820-14f0-0132-a142-66789f1cb1b5/status?branch=master)](https://codeship.io/projects/33558)


### Installation

Standard `go get`:

```
$ go get github.com/pabdavis/beego_keenio
```

#### Dependencies

```
  go get github.com/astaxie/beego 
  go get github.com/philpearl/keengo
```

### Usage

To use this beego middleware with the Keen IO API, you have to configure your Keen IO Project ID and its write access key (if you need an account, [sign up here](https://keen.io/) - it's free).

This configuration information needs to be added to the conf/app.conf file in your Beego project

```ini
 KeenioProjectId = XXXXXXXXX
 KeenioWriteKey =  YYYYYYYYY
```

Beware of whitespace and line breaks in the write key based on it's length.

#### Configuring Beego Middleware

Add the following lines into the ```routers/routers.go``` file which will initialize the filter to run on all requests (BeforeRouter and FinishRouter)


```go
 import "github.com/pabdavis/beego_keenio"
 
 func init() {
    beego_keenio.InitKeenioFilter()

  
 }
```

#### Queueing Events from controller

Since Keen IO does not force specific tags to be included, this middleware attempts to provide a flexible way for you to format the data you want 
to send and it will handle it from there. 

The filter will provide a empty queue via GetData which allows for multiple keen events per controller method to different event collections. 
Use the Push method to identify the Keen IO event collection and the data to send to the collection.  The data must be an interface that can 
be marshaled into JSON, sample uses simplejson. 

** You must set the variable back into input context using the beego_keenio constant, if not, the events will not be sent to keen io. 

 
```go

    func (this *Controller) SomeMethod() {
       
        apiData := map[string]interface{}{
            "apikey":   api.Key, 
            "app_name": api.Application.Name,
            "username": api.User.Name,
        }
        dataSet1 := simplejson.New()
        dataSet1.Set("api_request", apiData)
    
        ....

        purchaseData := map[string]interface{}{
            "item_id": item.Key
            "qty": 1
            "price": 5
        }
        dataSet2 := simplejson.New()
        dataSet2.Set("purchases", purchaseData)

        if keenQ, ok := this.Ctx.Input.GetData(beego_keenio.KEENIO_QUEUE_KEY).(beego_keenio.KeenioQueue); ok {
            keenQ.Push("collection1", dataSet1)
            keenQ.Push("collection2", dataSet2)
            this.Ctx.Input.SetData(beego_keenio.KEENIO_QUEUE_KEY, keenQ) //Must set this back into the defined key
        }

        ....
    }

```

That's it! After running your code, check your Keen IO Project to see the event/events has been added.
