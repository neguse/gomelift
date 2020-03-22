const io = require('socket.io')(3000);

io.on('connection', (socket) => {

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
    
})
