let standingsTable = document.getElementsByClassName("standings")[0];
let rows = Array.from(standingsTable.rows)

rows.forEach(row => {
  let deltaCell = document.createElement(row.cells[0].tagName)
  deltaCell.classList = row.cells[row.cells.length - 1].classList

  deltaCell.style = "width:4em;"

  if (deltaCell.tagName == "TH") {
    deltaCell.innerHTML = "&Delta;"
  } else if (!row.classList.contains("standingsStatisticsRow")) {
    deltaCell.innerHTML = "0"
  }

  if (row.cells.length > 1) {
    row.cells[row.cells.length - 1].remove("right")
    row.appendChild(deltaCell)
  }
})
