#Evroll
Evroll is a simple library combining a few techniques from dynamic languages like nodejs and standard golang inbuilt features to build interesting features like events,streamers and callback queues. 


##Install
    
    go get github.com/influx6/evroll
    
Then
    
    go install github.com/influx6/evroll

##API

###Rollers
Rollers are the standard callback queues powered by a buffer array underneath, it allows the standard node style callback based notifications for use where such pattern is feasibile. Rollers are built in a middleware style allowing control of the calling of the next callback within the list,it allows rollers to be used as standard callback queue chains or even middleware stack callback call chains.
        
- Evroll.NewRoller() *Roller
  
  Creates and returns a new Roller for use 

            ` 
                callbacks := evroll.NewRoller();
            
            `

- Evroll.Roller.Receive(func (i interface{})) void

    This receiver function allows the addition of a function matching the stated type into the callback queue, due to the nature of go, the value called on the function passed will be enclosed in the default interface{} type which all objects in go satisfy, since type is known to the user,simple user a type assertion to get the desired type.

            `
                callbacks := evroll.NewRoller();

                callbacks.Receive(func (data interface{}){
                    fmt.Println(“currently received data:”,data)
                })
            `
            
- Evroll.Roller.Decide(func (current interface{}, next func(newValue interface{}))) void

    The `Decide` receiver function of the `Roller` struct allows a more expanding normal middleware style control of the callback stack in the roller, it recieves both the data and a next function call that can take a value to be used as the new value for other callbacks until its is changed by another function call down the callback stack.

            `
                callbacks := evroll.NewRoller();

                callbacks.Decide(func (data interface{}, next func(newval interface{})){
                    fmt.Println(“currently received data:”,data)
                    next(nil)
                })
            `

- Evroll.Roller.RevMunch(data interface{}) void

    This reciever method initiates the data that’s sent into the callback chain which is propagated to all subscribed, because its type is the general interface{} type,any value can be sent in. The different between this method and the `Munch` method is that it reverses the callstack and calls in a LIFO order

            `
                callbacks := evroll.NewRoller();

                callbacks.Munch(300)

            `


- Evroll.Roller.Munch(data interface{}) void

    This reciever method initiates the data that’s sent into the callback chain which is propagated to all subscribed, because its type is the general interface{} type,any value can be sent in. This calles all the subscribers in a FIFO order.

            `
                callbacks := evroll.NewRoller();

                callbacks.Munch(300)

            `

- Evroll.Roller.Size() int

    This reciever method returns the current size of the total callbacks within the callback queue

            `
                callbacks := evroll.NewRoller();

                size := callbacks.Size()

            `

- Evroll.Roller.CallAt(index int,data interface{}) void

    This reciever method is the real worker,it cycles all callbacks from the given index and calls the data on each

            `
                callbacks := evroll.NewRoller();

                callbacks.CallAt(0,”buzz”)
            `

- Evroll.Roller.ReverseCallAt(index int,data interface{}) void

    This reciever method is the real worker,it cycles all callbacks in reverse from the given index and calls the data on each. It actually takes the index and increments and subtract it from the total length to get the correct index

            `
                callbacks := evroll.NewRoller();

                callbacks.CallAt(0,”buzz”)
            `

- Evroll.Roller.CallDoneAt(index int,data interface{}) void

    This reciever method is the real worker,it cycles all done callbacks from the given index and calls the data on each

            `
                callbacks := evroll.NewRoller();

                callbacks.CallDoneAt(0,”buzz”)
            `

- Evroll.Roller.ReverseCallDoneAt(index int,data interface{}) void

    This reciever method is the real worker,it cycles all done callbacks in reverse from the given index and calls the data on each. It actually takes the index and increments and subtract it from the total length to get the correct index

            `
                callbacks := evroll.NewRoller();

                callbacks.ReverseCallDoneAt(0,”buzz”)
            `


###Streams:
    
- Evroll.NewStream(reverse bool,manual bool) *Streams

This method returns a stream pointer and its a composition of the Roller struct to allow access to roller member methods. Streams where created to allow a simpler version without the standard issue of deadlock that can be heavily common with channels and was built with the desire to have it more functional. The reverse argument forces callbacks to be called manaually and the manaul forces pull like behaviour i.e until the ‘Stream()’ method is called nothing is done with the data and these evidently affects the drain notification behaviour,for when `manaul` is set to true, drain is called when no items remain in the streams buffer but when `manaul` is false, drain is practically called on every single instance of a `Send(data)` call because of the push behaviour
            

            `
                //creates a stream that ensures its callbacks are called in LIFO order
                streams := evroll.NewStream(false,false);

                //creates a stream that ensures its callbacks are called in FIFO order
                reverseStreams := evroll.NewStream(true,false);
            `

- Evroll.Streams.Drain(func (data interface{}))  void

    This reciever method hads a callback to the drain event handler and its called when all data in the stream has all been sent out to call listening callbacks

            `
                streams := evroll.NewStream(false);


                streams.Drain(func (data interface{}){
                    //do something...
                })

            `

- Evroll.Streams.Clear()  void

    This reciever method simply flushes the data in the stream buffer

            `
                streams := evroll.NewStream(false);

                streams.Clear()

            `

- Evroll.Streams.Send(data interface{})  void

    This reciever method queues up the data to be sent to all callbacks and depending on the bool value of `manaul` will immediately call the `Stream()` method or leave it to the caller if `manaul` is set to `true`

            `
                streams := evroll.NewStream(false);

                streams.Send(300)

            `

- Evroll.Streams.CollectAndStream()  void

    This reciever method collects all data within the streams and streams the whole list into itself (fun ehn....)

            `
                streams := evroll.NewStream(false);

                streams.CollectAndStream()

            `

- Evroll.Streams.CollectTo(func (data []interface{}))  void

    This reciever method collects all data within the streams and passes it to the provided function

            `
                streams := evroll.NewStream(false);

                streams.CollectTo(func (data []interface{}{
                    //do something...
                }))

            `

- Evroll.Streams.Collect()  []interface{}

    This reciever method collects all data within the streams and returns them as an array 

            `
                streams := evroll.NewStream(false);

                data := streams.Collect()

            `

- Evroll.Streams.Stream()  void

    This reciever method starts off the streaming of all data added into its buffer to call listening buffers,it exists to allow control of the pushing of the data down the train when needed when one decides not to allow push like effect on every data added but wishes to control by busting all the data after fully adding all data needed into the callback drain

            `
                streams := evroll.NewStream(false);

                streams.Send(400)
                streams.Stream()

            `


###Events
    
- Evroll.NewEvent(id string)  *EventRoll

    This creates a new `EventRoll` and returns the pointer to it

- Evroll.EventRoll.Listen(func (data interface{}))  void

    This reciever method hads a function into the listener list for the event roller

            `
                streams := evroll.NewEvent(“pack”);

                streams.Listen(func (data interface{}){
                    //do something...
                })

            `
- Evroll.EventRoll.Listen(func (data interface{}))  void

    This reciever method hads a function into the listener list for the event roller

            `
                streams := evroll.NewEvent(“pack”);

                streams.Listen(func (data interface{}){
                    //do something...
                })

            `

- Evroll.EventRoll.Emit(data interface{})  void

    This reciever method calls all added callbacks with the data provided by the caller

            `
                streams := evroll.NewEvent(“pack”);

                streams.Emit(400)

            `
