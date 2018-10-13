
function enqueueMessages(messages) {
  $("#messages").append($.map(messages, function(message) {
    return `<li>${message.Origin}: ${message.Text}</li>`;
  }));
}

$(function(){

  $("#send-message").click(function(){
    $.post('/message', JSON.stringify("Hello"), function(messages) {
      enqueueMessages(JSON.parse(messages));
    });
  });
});
