const process = require('node:process');
const lunr = require("./lunr")

process.stdin.resume()
process.stdin.setEncoding('utf8')

let buffer = []

process.stdin.on('data', function (data) {
    buffer.push(data)
})

process.stdin.on('end', function () {
    let items = JSON.parse(buffer.join(''));

    let idx = lunr(function () {
        this.ref('id')
        this.field('type')
        this.field('title')
        this.field('description')
        this.field('link')
        this.field('parent_id')

        for (let item of items) {
            this.add(item)
        }
    });

    process.stdout.write(JSON.stringify(idx))
})
