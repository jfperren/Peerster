
function enqueueMessages(messages) {
  $("#messages").append($.map(messages, function(message) {
    return `<li>${message.Origin}: ${message.Text}</li>`;
  }));
}

$(function(){


  $("#message-form").submit(function(){

    var message = $("#message").val();

    $.post('/message', JSON.stringify(message), function(messages) {
      var message = $("#message").val("");
      enqueueMessages(JSON.parse(messages));
    });
  });
});
