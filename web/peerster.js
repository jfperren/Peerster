
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
  }

  if (peers.includes(node)) {
    callback(null, `IP Address already in list of peers: ${node}`);
  }

  $.post('/node', JSON.stringify(node), function(res) {
    callback(JSON.parse(res), null);
  });
}

// --- DOM UPDATE --- //

function enqueueMessages(newMessages) {

  newMessages = newMessages.filter(function(message) {
    return !messages.includes(message);
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

function isValidIPAddress(address) {
  regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):[0-9]{4}$/;
  return regex.test(address)
}

// --- MAIN CODE --- //

var statuses = [];
var peers = [];
var messages = [];

$(function(){

  $.get("/id", function(res) {
    var title = JSON.parse(res)
    $("#node-title").html(title)
  });

  loadNewPeers()
  loadNewMessages()

  setInterval(loadNewMessages, 1000)
  setInterval(loadNewPeers, 1000)

  $("#peer-form").submit(function(){

    // Get the peer address, exit if it's empty
    var peer = $("#peer").val()
    if (peer == "") { return }

    postNode(peer, function(res, err) {

      // If there's an error, alert and exit
      if (err != null) { alert(err); return }

      // Reset value in field
      $("#peer").val("")

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
});
