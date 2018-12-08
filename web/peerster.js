
// --- SERVICE LAYER --- //

function getMessages(statuses, callback) {
  $.ajax({
    method: "GET",
    url: "/message",
    headers: {
      'x-statuses': JSON.stringify(statuses)
    },
    success: function(res) {
      callback(JSON.parse(res));
    }
  });
};

function postMessage(message, statuses, callback) {
  $.ajax({
    method: "POST",
    url: "/message",
    data: JSON.stringify(message),
    headers: {
      'x-statuses': JSON.stringify(statuses)
    },
    success: function(res) {
      console.log(res)
      callback(JSON.parse(res));
    },
    error: function(res) {
      console.log(res)
      callback(null, JSON.parse(res));
    }
  });
};

function getNodes(callback) {
  $.get("/node", function(res) {
    callback(JSON.parse(res), null);
  });
}

function postNode(node, callback) {

  if (!isValidIPAddress(node)) {
    callback(null, `Invalid IP Address: ${node}`);
    return
  }

  if (peers.includes(node)) {
    callback(null, `IP Address already in list of peers: ${node}`);
    return
  }

  $.post('/node', JSON.stringify(node), function(res) {
    callback(JSON.parse(res), null);
  });
}

function getUsers(callback) {
    $.get("/user", function(res) {
        callback(JSON.parse(res), null);
    });
}

function postPrivateMessage(message, destination, callback) {
    $.ajax({
        method: "POST",
        url: "/privateMessage",
        data: JSON.stringify({ 'Destination': destination, 'Text': message }),
        success: function(res) {
            callback(JSON.parse(res));
        }
    });
};

function getPrivateMessages(callback) {
  $.ajax({
    method: "GET",
    headers: {
      'x-index': privateMessages.length
    },
    url: "/privateMessage",
    success: function(res) {
      callback(JSON.parse(res));
    },
    error: function(res) {
      callback(null, JSON.parse(res));
    }
  });
};

function uploadFile(filename, callback) {

  if ($.map(files, (f) => f.Name).includes(filename)) {
    callback(null, `$filename is already shared onto the network.`)
  }

  $.ajax({
    method: "POST",
    url: "/fileUpload",
    data: JSON.stringify(filename),
    success: function(res) {
      callback(JSON.parse(res));
    },
    error: function(res) {
      console.log(res)
      callback(null, res);
    }
  });
};

function searchRequest(keywords, callback) {
  $.ajax({
    method: "POST",
    url: "/fileSearch",
    data: JSON.stringify({
      'Keywords': keywords,
      'Budget': 4
    }),
    success: function(res) {
      callback(JSON.parse(res));
    },
    error: function(res) {
      callback(null, res);
    }
  });
}

function getUploadedFiles(callback) {

  $.get("/fileUpload", function(res) {
    callback(JSON.parse(res));
  });
}

function getSearchResult(callback) {

  $.get("/fileSearch", function(res) {
    callback(JSON.parse(res));
  });
}

function downloadFilePrivate(filename, hash, destination, callback) {

  if ($.map(files, (f) => f.Name).includes(filename)) {
    callback(null, `$filename is already in your list of files.`)
  }

  $.ajax({
    method: "POST",
    url: "/fileDownload",
    data: JSON.stringify({
      'Destination': destination,
      'Hash': hash,
      'Name': filename
    }),
    success: function(res) {
      callback(JSON.parse(res));
    },
    error: function(res) {
      callback(null, res);
    }
  });
};

function downloadFilePublic(filename, hash, callback) {

  if ($.map(files, (f) => f.Name).includes(filename)) {
    callback(null, `$filename is already in your list of files.`)
  }

  $.ajax({
    method: "POST",
    url: "/fileDownload",
    data: JSON.stringify({
      'Name': filename,
      'Hash': hash,
    }),
    success: function(res) {
      callback(JSON.parse(res));
    },
    error: function(res) {
      callback(null, res);
    }
  });
};

// --- DOM UPDATE --- //

function enqueueMessages(newMessages) {

  newMessages = newMessages.filter(function(message) {
      return !messages.includes(message) && message.Text != "";
  })

  messages = messages.concat(newMessages);

  $("#messages").append($.map(newMessages, function(message) {
      if (message.Origin == name) {
        return `<li class="shout" >\<<span style='font-weight:400;'>${message.Origin} (You)</span>\> ${message.Text}</li>`;
      } else {
        return `<li class="shout">\<<a class="send-private" href="#" to="${message.Origin}">${message.Origin}</a>\> ${message.Text}</li>`;
      }
  }));
}

function enqueuePrivateMessages(newPrivateMessages) {

  newPrivateMessages = newPrivateMessages.filter(function(privateMessage) {
    return !privateMessages.includes(privateMessage);
  })

  privateMessages = privateMessages.concat(newPrivateMessages);

  $("#messages").append($.map(newPrivateMessages, function(privateMessage) {

      if (privateMessage.Origin == name) { // Message from us
          message = `[ PM to <a class="send-private" href="#" to="${privateMessage.Destination}">${privateMessage.Destination}</a> ] ${privateMessage.Text}`
      } else { // Message for us
          message = `[ PM from <a class="send-private" href="#" to="${privateMessage.Origin}">${privateMessage.Origin}</a> ] ${privateMessage.Text}`
      }

      return `<li class="whisper">${message}</li>`;
  }));
}

function enqueuePeers(newPeers) {

    newPeers = newPeers.filter(function(peer) {
        return !peers.includes(peer);
    })

    peers = peers + newPeers;

    $("#peers").append($.map(newPeers, function(peer) {
        return `<li>${peer}</li>`
    }));
}

function enqueueUsers(newUsers) {

    newUsers = newUsers.filter(function(user) {
        return !users.includes(user);
    })

    users = users + newUsers;

    $("#users").append($.map(newUsers, function(user) {
        return `<li><a class="send-private" href="#" to="${user.Name}">${user.Name}</a> (${user.Address})</li>`
    }));
}

function enqueueFiles(newFiles) {

    newFiles = newFiles.filter(function(file) {
        return !files.includes(file);
    })

    files = files + newFiles

    $("#files").append($.map(newFiles, function(file) {
      return `<li>${file.Name} (<span class="hash">${file.Hash}</span>)</li>`;
    }));
}

function enqueueSearchResults(results) {

    newResults = results.filter(function(result) {
        return !searchResults.includes(result);
    })

    searchResults = searchResults + newResults;

    $("#search-results").append($.map(newResults, function(result) {

        html = `<li><div class="search-result">${result.Name}</div>`
        html += `<span class="hash">`

        if (result.Full) {
          html += `> <a class="download-public" href="#" hash="${result.Hash}">${result.Hash}</a>`
        } else {
          html += `> Download unavailable, some chunks are missing`
        }

        html += `</span></li>`

        return html
    }));
}

// --- CONVENIENCE METHODS --- //

function loadNewMessages() {
    getMessages(statuses, function(res) {
      enqueueMessages(res["Rumors"])
      statuses = res["Statuses"]
    });
}

function loadNewPrivateMessages() {
  getPrivateMessages(function(res) {
    enqueuePrivateMessages(res)
  });
}

function loadNewPeers() {
  getNodes(function(res, err) {
    enqueuePeers(res)
  });
}

function loadNewUsers() {
  getUsers(function(res, err) {
      enqueueUsers(res)
  });
}

function loadUploadedFiles() {
  getUploadedFiles(function(res, err) {
    enqueueFiles(res)
  });
}

function loadSearchResults() {
  getSearchResult(function(res, err) {
    enqueueSearchResults(res)
  });
}

function isValidIPAddress(address) {
  regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):[0-9]{4}$/;
  return regex.test(address)
}

// --- MAIN CODE --- //

var statuses = [];
var peers = [];
var messages = [];
var privateMessages = [];
var users = [];
var name = "";
var files = [];
var searchResults = [];

$(function(){

  $.get("/id", function(res) {
    var title = JSON.parse(res)
    name = title
    $("#node-title").html(title)
    loadNewPeers()
    loadNewMessages()
    loadNewUsers()
    loadNewPrivateMessages()
    loadUploadedFiles()
    loadSearchResults()
  });

  setInterval(loadNewMessages, 1000)
  setInterval(loadNewPeers, 1000)
  setInterval(loadNewUsers, 1000)
  setInterval(loadNewPrivateMessages, 1000)
  setInterval(loadSearchResults, 1000)

  $("#add-peer").on('click', function(e){
    e.preventDefault();

    // Get the peer address, exit if it's empty
    var peer = prompt(`Connect to a new peer address`, "127.0.0.1:0000")

    if (peer == null || peer == "") {
      return
    }

    postNode(peer, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { alert(err); return }

      // Add peer to list
      enqueuePeers([peer])
    });
  });

  $("#message-form").submit(function(){

    // Get the message, exit if it's empty
    var message = $("#message").val()
    if (message == "") { return }

    postMessage(message, statuses, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { alert(err); return }

      // Reset value in field
      $("#message").val("");

      // Enqueue new messages
      enqueueMessages(res["Rumors"]);

      // Updates status
      statuses = res["Statuses"];
    });
  });

  $("body").on('click', '.send-private', function(e) {
    e.preventDefault();

    to = $(e.target).attr('to')

    if (to == null || to == "") {
      console.log("Error, 'to' should be defined for .send-private objects")
      return
    }

    var message = prompt(`Send a private message to ${to}`, "Your message here...")

    if (message == null || message == "") {
      return
    }

    postPrivateMessage(message, to, function(res) { });
  });

  $("#upload-file").on('click', function(e){
    e.preventDefault();

    // Get the peer address, exit if it's empty
    var filename = prompt("Enter the file name", "myFile.txt")

    if (filename == null || filename == "") {
      return
    }

    uploadFile(filename, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { alert(err.responseText); return }

      // Add peer to list
      enqueueFiles([res])
    });
  });

  $("#download-file").on('click', function(e){
    e.preventDefault();

    var hash = prompt("What is the hash of the file you would like to download?",
      "8b786a45dcdb2db3f321f9cbe5d1126ebc7fd6b4bf5813fb4e8ae8fe6827daf5")

    if (hash == null || hash == "") {
      return
    }

    var destination = prompt("Who owns this file?", "Alice")

    if (destination == null || destination == "") {
      return
    }

    var filename = prompt("How would you like to name this file?", "myFile.txt")

    if (filename == null || filename == "") {
      return
    }

    downloadFilePrivate(filename, hash, destination, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { alert(err.responseText); return }

      // Add peer to list
      enqueueFiles([res])
    });
  });

  $("#search-form").submit(function(){

    // Get the message, exit if it's empty
    var query = $("#search-query").val()
    if (query == "") { alert("Please enter a least one character in search");return }

    searchRequest(query, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { console.log(err); return }

      // Reset value in field
      $("#search-query").val("");
    });
  });

  $("body").on('click', '.download-public', function(e) {
    e.preventDefault();

    hash = $(e.target).attr('hash')

    var filename = prompt("How would you like to name this file locally?")

    if (filename == null || filename == "") {
      return
    }

    downloadFilePublic(filename, hash)
  });
});
