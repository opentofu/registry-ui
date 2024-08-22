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
    let idx = generateIndex(items)
    process.stdout.write(JSON.stringify(idx))
})

function generateIndex(items) {
    let idx = lunr(function () {
        this.ref('ref')
        this.field('type')
        this.field('title', { boost: 10 })
        this.field('description')
        this.field('link')

        for (let item of items) {
            item.ref = getRef(item)
            this.add(item)
        }
    });

    return idx
}

// getRef returns the ref for the item, this is the only information returned when searching
// so it must contain all the information we need to display the item
// for this reason, we will be json serializing what we need to put in there
function getRef(item) {
    return JSON.stringify({
        addr: item.addr,
        type: item.type,
        version: item.version,
        title: item.title,
        description: item.description,
        link: item.link,
        parent_id: item.parent_id
    })
}
