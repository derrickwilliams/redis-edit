redata = ((F) => {
  const keysPath = "/api/cache/keys";
  const getEntryPath = "/api/cache/get";
  const setEntryPath = "/api/cache/set";

  return {
    getKeys: () => {
      return F(`${keysPath}`)
        .then(response => response.json())
        .then(json => json.data)
        .catch(err => console.log('error', err))
    },
    getEntry: key => {
      return F(`${getEntryPath}/${key}`)
        .then(response => response.text())
        .catch(err => console.log('error', err))
    },
    updateEntry: (key, entry) => {
      return F(`${setEntryPath}/${key}`, { method: 'POST', body: entry })
        .then(response => response.json())
        .then(json => json.data)
        .catch(err => console.log('error', err))
    }
  }
})(fetch)

reDatastore = (() => {
  let _data = {}
  return (key, value) => {
    _data[key] = value
  }
})()

reBootstrap = ((global, redata, datastore, CM, $v) => {
  let editorBox;

  document.getElementById('save-btn').addEventListener('click', () => {
    if (confirm("Are you sure?")) {
      const list = document.getElementById('key-list')
      const selected = list.value
      redata.updateEntry(selected, editorBox.getValue())
        .then(results => alert(`${selected} value was saved`))
    }
  })

  document.getElementById('load-entry').addEventListener('click', () => {
    const list = document.getElementById('key-list')
    const selected = list.value
    redata.getEntry(selected)
      .then(entry => {
        reDatastore(`${selected}.pristine`, entry)
        return entry
      })
      .then(entry => initEditor(entry))
  })

  function initEditor(entry) {
    const box = document.getElementById('editor-box')
    box.innerText = entry

    if (editorBox !== undefined) {
      editorBox.setValue(entry);
      return editorBox
    }

    editorBox = CM.fromTextArea(box, {
      lineNumbers: true,
      mode: "xml",
      theme: "dracula",
      foldGutter: { scanUp: true, minFoldSize: 1 },
      matchTags: { bothTags: true },
      gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"],
      extraKeys: {"Ctrl-J": "toMatchingTag", "Ctrl-Q": function(cm){ cm.foldCode(cm.getCursor()); }}
    });

    return editorBox;
  }

  function renderKeysList(keys) {
    const list = document.getElementById('key-list')
    keys.forEach((k, i) => {
      opt = document.createElement('option')
      opt.text = k
      opt.value = k
      list.add(opt, i)
    })
  }

  return function bootstrapper() {

    redata.getKeys()
      .then(keys => {
        datastore(`keys`, keys)
        return keys
      })
      .then(keys => renderKeysList(keys))
      .catch(err => console.log('error', err))
  }
})(window, redata, reDatastore, CodeMirror, verge);