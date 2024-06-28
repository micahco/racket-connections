const toggleChecked = (nodes) => {
    let allChecked = true
    for (const td of nodes) {
        if (!td.querySelector("input").checked) {
            allChecked = false
            break
        }
    }

    for (const td of nodes) {
        td.querySelector("input").checked = allChecked ? false : true
    }
}

const rowNodes = (th) => {
    const nodes = []

    let cur = th
    while (cur.nextElementSibling) {
        cur = cur.nextElementSibling
        nodes.push(cur)
    }

    return nodes
}

const colNodes = (col) => {
    const nodes = []

    const rows = document.querySelector("tbody").getElementsByTagName("tr")
    for (const row of rows) {
        const td = row.getElementsByTagName("td")[col]
        nodes.push(td)
    }

    return nodes
}

for (const th of document.querySelectorAll("th[scope='row']")) {
    th.querySelector("button").addEventListener("click", e => {
        toggleChecked(rowNodes(th))
    })
}

const colThs = document.querySelectorAll("th[scope='col']")
for (let i = 1; i < colThs.length; i++) {
    colThs[i].querySelector("button").addEventListener("click", e => {
        toggleChecked(colNodes(i - 1))
    })
}
