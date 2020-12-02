// set constants and required packages
const express = require("express");
const app = express();
const port = 8080;

const bodyParser = require("body-parser");
const path = require("path");
const createError = require("http-errors");
const { slackMessageHandler, slackInteractiveHandler } = require("./internal/routes/slackHandler");

// Configure middleware
app.set("port", process.env.PORT || port); // set express to use this port
app.set("view engine", "ejs"); // configure template engine
app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json()); // parse form data client

// Application routes
//app.get("/migrate/guides", getGuideMigrationPage);

// Default response
app.get("/", (req, res, next) => {
	res.send("Hello, world!");
});

// Incoming slack message handler
app.get("/slackMessage", slackMessageHandler);
app.post("/slackMessage", slackMessageHandler);

// Incoming slack interactive element handler
// Incoming slack message handler
app.get("/slackInteractive", slackInteractiveHandler);
app.post("/slackInteractive", slackInteractiveHandler);

// 404 Handler (runs if page does not match any other routes)
app.use(function (req, res, next) {
    res.status(404);
	return next(createError(404, "Page Not Found"));
});

// Generic error handler
app.use(function errorHandler(err, req, res, next) {
	return next(err);
});

// launch app
app.listen(process.env.PORT || port, () =>
	console.log(`Server running on port: ${process.env.PORT || port}`)
);