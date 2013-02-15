var MAX_SCROLLBACK = 100;
var ws = new WebSocket("ws://localhost:8080/ws"); 

$(document).ready(function() {
	ws.onmessage = function(rawMsg) {
		var msg = JSON.parse(rawMsg.data)
		writeMessage(msg.Command + ": " + msg.Data)
	};
	
	$("#textEntry").keyup(function(evt) {
		if (evt.keyCode == 13 && !evt.shiftKey) {
			var textEntry = $(this); 
			var msg = textEntry.val().trim();
			
			var firstSpace = msg.indexOf(" ");
			
			if (firstSpace == -1) {
				var jsonMsg = JSON.stringify({"Command": msg});
			} else {
				var jsonMsg = JSON.stringify({
					"Command": msg.substring(0,firstSpace),
					"Data" : msg.substring(firstSpace+1)
				});
			}
			
			var success = ws.send(jsonMsg);
			
			textEntry.val("");
			
			if (!success) {
				writeMessage("Send failed, trying to reconnect...");
				connect();
			}
		}
	});
});

function writeMessage(msg) {
	var chat = $("#chat");
	
	chat.append($("<pre>").text(msg));
	
	if (chat.children().length > MAX_SCROLLBACK) {
		$("#chat :first-child").remove();
	}
	
	chat.animate({scrollTop: chat[0].scrollHeight});   
}
