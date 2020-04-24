async function getRatingChanges(contestId) {
  let server = (await browser.storage.sync.get("server")).server || 'http://localhost:8080'

  let url = `${server}/api/contest.ratingChanges?contestId=${contestId}`
  let res = await (await fetch(url)).json()
  return res.result
}

let contestId = window.location.href.match(/codeforces\.com\/contest\/(\d+)/)[1]

getRatingChanges(contestId).then(ratingChanges => {
  let deltas = {}
  ratingChanges.forEach(ratingChange => {
    deltas[ratingChange.handle] = ratingChange.newRating - ratingChange.oldRating
  })

  let standingsTable = document.getElementsByClassName("standings")[0]
  let rows = Array.from(standingsTable.rows)

  rows.forEach(row => {
    let deltaCell = document.createElement(row.cells[0].tagName)
    deltaCell.classList = row.cells[row.cells.length - 1].classList

    deltaCell.style = "width:4em;"

    if (deltaCell.tagName == "TH") {
      deltaCell.innerHTML = "&Delta;"
    } else if (!row.classList.contains("standingsStatisticsRow")) {
      let handle = row.cells[1].getElementsByClassName("rated-user")[0].innerHTML
      if (deltas[handle] !== undefined) {
        deltaCell.innerHTML = (deltas[handle] > 0 ? '+' : '') + deltas[handle].toString()

        if (deltas[handle] > 0) {
          deltaCell.style = "font-weight: bold; color: green;"
        } else {
          deltaCell.style = "font-weight: bold; color: gray;"
        }
      }
    }

    if (row.cells.length > 1) {
      row.cells[row.cells.length - 1].remove("right")
      row.appendChild(deltaCell)
    }
  })
}).catch(console.error)
