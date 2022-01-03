'use strict';

const MOVEMENT_SPEED = 10
const ZOOM_SPEED = 10
const POPUP_WIDTH = 500
const POPUP_HEIGHT = 370

var conn
var flashStatus = false

var focus = ""
var graph = {}
var root

var xpos = 0, ypos = 0
var scale = 100
var distance = 500
var depth = 2

var mousex, mousey
var keys = []

const titleMessage = document.getElementById("titlemessage");
const canvas = document.getElementById("graphcanvas");
canvas.width = window.innerWidth
canvas.height = window.innerHeight
const ctx = canvas.getContext("2d");

const plusButton = document.getElementById("plus");
const minusButton = document.getElementById("minus");
const pauseButton = document.getElementById("pause");
const continueButton = document.getElementById("continue");
const depthNumberText = document.getElementById("depth");
const maxDepthNumberText = document.getElementById("maxdepth");
const statusText = document.getElementById("above-bottom-buttons");

window.onload = function () {

	if (window["WebSocket"]) {

		conn = new WebSocket("ws://" + document.location.host);
		conn.onmessage = function (evt) {
			const data = JSON.parse(evt.data)
			if (data.root !== undefined) {
				root = data.root
				titleMessage.innerHTML = root.login
				root.avatar = new Image
				root.avatar.src = root.avatar_url
				statusText.innerHTML = root.login + " connected!"
			} else if (data.paused !== undefined) {
				if (data.paused) pauseButton.style.background = "grey";
				else             pauseButton.style.background = "";
				if (!data.paused) continueButton.style.background = "grey";
				else              continueButton.style.background = "";
				depthNumberText.innerHTML = data.depth
				if (data.working) {
					if (statusText.innerHTML !== "Wrapping up...") {
						statusText.innerHTML = "Fetching user data..."
						flashStatus = true
					}
				} else if (data.paused) {
					flashStatus = false
					statusText.innerHTML = "Searching paused."
				} else {
					statusText.innerHTML = ""
				}
			} else if (data.username !== undefined) {
				graph[data.username] = data.collaborators
				data.collaborators.forEach(c => {
					c.avatar = new Image
					c.avatar.src = c.avatar_url
				})
			}
		};

		plusButton.onclick = () => { depth++ };
		minusButton.onclick = () => { if (--depth < 0) depth = 0; };
		pauseButton.onclick = () => { statusText.innerHTML = "Wrapping up..."; conn.send(JSON.stringify({command:"pause"})); };
		continueButton.onclick = () => { conn.send(JSON.stringify({command:"continue"})); };
		document.addEventListener('keydown', (evt) => { keys[evt.keyCode] = true; })
		document.addEventListener('keyup',   (evt) => { keys[evt.keyCode] = false; })
		window.addEventListener('mousemove', mousecapture, false);

		setInterval(update, 1000/30);
		setInterval(toggleStatusText, 1000/2);

	} else {
		titleMessage.innerHTML = "<b>Your browser does not support WebSockets.</b>";
	}
}

function drawLine(x1, y1, x2, y2) {
	ctx.strokeStyle = 'white';
	ctx.lineWidth = 1;
	ctx.beginPath();
	ctx.moveTo(x1, y1);
	ctx.lineTo(x2, y2);
	ctx.stroke();
}

function drawAvatar(node, x, y, s) {
	ctx.beginPath();
    ctx.arc(x, y, s + 1, 0, Math.PI * 2, true);
	ctx.fillStyle = 'white';
	ctx.fill();
	ctx.save();
	ctx.beginPath();
	ctx.arc(x, y, s, 0, Math.PI * 2, true);
	ctx.closePath();
	ctx.clip();
	ctx.drawImage(node.avatar, x - s, y - s, s * 2, s * 2);
	ctx.beginPath();
	ctx.arc(x, y, s, 0, Math.PI * 2, true);
	ctx.clip();
	ctx.closePath();
	ctx.restore();
}

function drawGraph(parent, x, y, s, d, r) {
	let children = graph[parent.login]
	if (r > 0 && children) {
		for (let i = 0; i < children.length; i++) {
			let child_x = x + Math.cos(2 * Math.PI * i / children.length) * d
			let child_y = y + Math.sin(2 * Math.PI * i / children.length) * d
			drawLine(x, y, child_x, child_y);
			drawGraph(children[i], child_x, child_y, s / 2, d / 2, r - 1)
		}
	}
	drawAvatar(parent, x, y, s)
}

function drawPopup(parent, x, y, s, d, r) {
	if (Math.abs(mousex - x) < s && Math.abs(mousey - y) < s) {
		let tx = mousex
		let ty = mousey
		if (tx > canvas.width - POPUP_WIDTH) tx -= POPUP_WIDTH
		if (ty > canvas.height - POPUP_HEIGHT) ty -= POPUP_HEIGHT

		ctx.fillStyle = "white";
		ctx.fillRect(tx, ty, POPUP_WIDTH, POPUP_HEIGHT);

		ctx.fillStyle = "black";
		ctx.textAlign = "left"
		ctx.font = "bold 38px helvetica";
		ctx.fillText(parent.login, tx + 20, ty + 50, POPUP_WIDTH - 40);
		ctx.font = "20px helvetica";
		ctx.fillText(parent.name || "", tx + 20, ty + 80, POPUP_WIDTH - 40);

		ctx.fillText(parent.location || "--", tx + 20, ty + 125, POPUP_WIDTH / 2 - 20);
		ctx.fillText(parent.company || "--", tx + 20, ty + 150, POPUP_WIDTH / 2 - 20);
		ctx.fillText(parent.email || "--", tx + POPUP_WIDTH / 2 + 10, ty + 125, POPUP_WIDTH / 2 - 20);
		ctx.fillText(parent.blog || "--", tx + POPUP_WIDTH / 2 + 10, ty + 150, POPUP_WIDTH / 2- 20);

		let age = "Unknown"
		if (parent.created_at) {
			age = timeSince(new Date(parent.created_at))
		}

		if (parent == root) {
			ctx.fillText(parent.public_repos || "Unknown", tx + 20, ty + 235, POPUP_WIDTH / 2 - 20);
			ctx.fillText(age, tx + 20, ty + 295, POPUP_WIDTH / 2 - 20);
			ctx.fillText(parent.followers || "Unknown", tx + POPUP_WIDTH / 2 + 10, ty + 235, POPUP_WIDTH / 2 - 20);
			ctx.fillText(parent.following || "Unknown", tx + POPUP_WIDTH / 2 + 10, ty + 295, POPUP_WIDTH / 2 - 20);
		} else {
			ctx.fillStyle = (parent.public_repos && root.public_repos && parent.public_repos != root.public_repos) ? ((parent.public_repos > root.public_repos) ? "green" : "red") : "black"
			ctx.fillText(parent.public_repos || "Unknown", tx + 20, ty + 235, POPUP_WIDTH / 2 - 20);
			ctx.fillStyle = (parent.created_at && root.created_at && parent.created_at != root.created_at) ? ((new Date(parent.created_at) > new Date(root.created_at)) ? "green" : "red") : "black"
			ctx.fillText(age, tx + 20, ty + 295, POPUP_WIDTH / 2 - 20);
			ctx.fillStyle = (parent.followers && root.followers && parent.followers != root.followers) ? ((parent.followers > root.followers) ? "green" : "red") : "black"
			ctx.fillText(parent.followers || "Unknown", tx + POPUP_WIDTH / 2 + 10, ty + 235, POPUP_WIDTH / 2 - 20);
			ctx.fillStyle = (parent.following && root.following && parent.following != root.following) ? ((parent.following > root.following) ? "green" : "red") : "black"
			ctx.fillText(parent.following || "Unknown", tx + POPUP_WIDTH / 2 + 10, ty + 295, POPUP_WIDTH / 2 - 20);
		}

		ctx.fillStyle = "black";
		ctx.font = "bold 20px helvetica";
		ctx.fillText("Public repos", tx + 20, ty + 210, POPUP_WIDTH / 2 - 40);
		ctx.fillText("Account Age", tx + 20, ty + 270, POPUP_WIDTH / 2 - 40);
		ctx.fillText("Followers", tx + POPUP_WIDTH / 2 + 10, ty + 210, POPUP_WIDTH / 2 - 40);
		ctx.fillText("Following", tx + POPUP_WIDTH / 2 + 10, ty + 270, POPUP_WIDTH / 2 - 20);

		ctx.font = "italic 20px serif";
		ctx.textAlign = "center"
		if (parent.bio)
			ctx.fillText(parent.bio || "", tx + POPUP_WIDTH / 2, ty + 350, POPUP_WIDTH - 40);

	} else {
		let children = graph[parent.login]
		if (r > 0 && children) {
			for (let i = 0; i < children.length; i++) {
				let child_x = x + Math.cos(2 * Math.PI * i / children.length) * d
				let child_y = y + Math.sin(2 * Math.PI * i / children.length) * d
				drawPopup(children[i], child_x, child_y, s / 2, d / 2, r - 1)
			}
		}
	}
}

function update() {
	maxDepthNumberText.innerHTML = depth

	// Keyboard control
	if (keys[37]) { xpos += MOVEMENT_SPEED }
	if (keys[38]) { ypos += MOVEMENT_SPEED }
	if (keys[39]) { xpos -= MOVEMENT_SPEED }
	if (keys[40]) { ypos -= MOVEMENT_SPEED }
	//if (keys[187]) { scale += ZOOM_SPEED; distance += ZOOM_SPEED * 2}
	//if (keys[189] && scale > 50) { scale -= ZOOM_SPEED; distance -= ZOOM_SPEED * 2}

	ctx.clearRect(0, 0, canvas.width, canvas.height);
	if (root === undefined || !graph[root.login])
		return
	drawGraph(
		root, xpos + canvas.width / 2, ypos + canvas.height / 2,
		scale, distance, depth)

	drawPopup(
		root, xpos + canvas.width / 2, ypos + canvas.height / 2,
		scale, distance, depth)
}

function mousecapture(evt) {
	var rect = canvas.getBoundingClientRect();
	mousex = (evt.clientX - rect.left) / (rect.right - rect.left) * canvas.width
	mousey = (evt.clientY - rect.top) / (rect.bottom - rect.top) * canvas.height
}

function toggleStatusText() {
	if (statusText.style.display === "none")
		statusText.style.display = "block"
	else if (flashStatus)
		statusText.style.display = "none"
}

function timeSince(date) {
	var seconds = Math.floor((new Date() - date) / 1000);
	var interval = seconds / 31536000;

	if (interval > 1) {
		return Math.floor(interval) + " years";
	}
	interval = seconds / 2592000;
	if (interval > 1) {
		return Math.floor(interval) + " months";
	}
	interval = seconds / 86400;
	if (interval > 1) {
		return Math.floor(interval) + " days";
	}
	interval = seconds / 3600;
	if (interval > 1) {
		return Math.floor(interval) + " hours";
	}
	interval = seconds / 60;
	if (interval > 1) {
		return Math.floor(interval) + " minutes";
	}

	return Math.floor(seconds) + " seconds";
}
