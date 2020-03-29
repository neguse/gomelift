const io = require('socket.io')(3000);

io.on('connection', (socket) => {
    console.log('connection', socket);

    socket.emit('ccc', (data) => {
        console.log('ccc', data)
    })

    socket.on("aaa", (data, callback) => {
        console.log(data);
        callback('bbb');
        socket.emit('fff', (data) => {
            console.log(data)
        })
    });

    socket.on('error', (error) => {
        console.log(error);
    });

    socket.on('disconnecting', (reason) => {
        console.log(reason);
    });
    socket.on('disconnect', (reason) => {
        console.log(reason);
    });
    
})
