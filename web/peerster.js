
$(function(){

  $("#peer-form").submit(function(){

    var newPeer = $("#peer").val()

    $.post('/node', JSON.stringify(newPeer), function(res) {

      // Test error -> Alert

      $("#peer").val("");
      var peer = JSON.parse(res);
      $("#peers").append(`<li>${peer}</li>`);
    });
  });

  $("#message-form").submit(function(){

    var message = $("#message").val()

    if (message != "") {
      $.post('/message', JSON.stringify(message), function(messages) {
        $("#message").val("");
        var newMessages = JSON.parse(messages);

        console.log(newMessages)

        $("#messages").append($.map(newMessages, function(message) {
          return `<li>${message.Origin}: ${message.Text}</li>`;
        }));
      });
    }
  });
});
