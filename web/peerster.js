
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
      callback(JSON.parse(res));
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

// --- DOM UPDATE --- //

function enqueueMessages(newMessages) {

  newMessages = newMessages.filter(function(message) {
    return !messages.includes(message) && message.Text != "";
  })

  messages = messages.concat(newMessages);

  $("#messages").append($.map(newMessages, function(message) {
    return `<li>${message.Origin}: ${message.Text}</li>`;
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
        return `<li><a class="send-private" href="#">${user.Name} @ ${user.Address}</a></li>`
    }));
}

// --- CONVENIENCE METHODS --- //

function loadNewMessages() {
  getMessages(statuses, function(res) {
    enqueueMessages(res["Rumors"])
    statuses = res["Statuses"]
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

function isValidIPAddress(address) {
  regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):[0-9]{4}$/;
  return regex.test(address)
}

// --- MAIN CODE --- //

var statuses = [];
var peers = [];
var messages = [];
var users = [];

$(function(){

  $.get("/id", function(res) {
    var title = JSON.parse(res)
    $("#node-title").html(title)
  });

  loadNewPeers()
  loadNewMessages()
  loadNewUsers()

  setInterval(loadNewMessages, 1000)
  setInterval(loadNewPeers, 1000)
  setInterval(loadNewUsers, 1000)

  $("#add-peer").on('click', function(e){
    e.preventDefault();

    // Get the peer address, exit if it's empty
    var peer = prompt(`Connect to a new peer address`, "127.0.0.1:0000")

    if (peer == null || peer == "") {
      return
    }

    // var peer = $("#peer").val()
    // if (peer == "") { return }

    postNode(peer, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { alert(err); return }

      // Reset value in field
      // $("#peer").val("")

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

  $("#users").on('click', '.send-private', function(e) {
    e.preventDefault();

    var message = prompt(`Send a private message to {"Alice"}`, "Your message here...")

    if (message == null || message == "") {
      return
    }

    postPrivateMessage(message, "Alice", function(res) { });
  });
});
