

var statuses = [];

function enqueueMessages(messages) {
  $("#messages").append($.map(messages, function(message) {
    return `<li>${message.Origin}: ${message.Text}</li>`;
  }));
}

function enqueuePeers(peers) {
  $("#peers").append($.map(peers, function(peer) {
    return `<li>${peer}</li>`
  }));
}

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

function loadNewMessages() {
  getMessages(statuses, function(res) {
    enqueueMessages(res["Rumors"])
    statuses = res["Statuses"]
  });
}



$(function(){

  $.get("/id", function(res) {
    var title = JSON.parse(res)
    $("#node-title").html(title)
  });

  $.get("/node", function(res) {
    var peers = JSON.parse(res)
    enqueuePeers(peers)
  });

  loadNewMessages()

  setInterval(loadNewMessages, 1000)

  $("#peer-form").submit(function(){
    var newPeer = $("#peer").val()
    $.post('/node', JSON.stringify(newPeer), function(res) {
      $("#peer").val("");
      var peer = JSON.parse(res);

      if (peer != "") {
        enqueuePeers([peer])
      }
    });
  });

  $("#message-form").submit(function(){

    var message = $("#message").val()

    if (message == "") {
      return
    }

    postMessage(message, statuses, function(res) {
      $("#message").val("");
      enqueueMessages(res["Rumors"])
      statuses = res["Statuses"]
    });
  });
});
