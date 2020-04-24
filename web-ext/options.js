function saveOptions(e) {
  let server = document.querySelector("#server").value
  let permissions = { origins: [`${server}/*`] }

  browser.storage.sync.set({ server })
  browser.permissions.request(permissions)
    .then((res) => {
      console.log(res)
    })

  e.preventDefault()
}

function restoreOptions() {
  let gettingItem = browser.storage.sync.get('server')
  gettingItem.then((res) => {
    document.querySelector("#server").value = res.server || 'http://localhost:8080'
  })
}

document.addEventListener('DOMContentLoaded', restoreOptions)
document.querySelector("form").addEventListener("submit", saveOptions)
