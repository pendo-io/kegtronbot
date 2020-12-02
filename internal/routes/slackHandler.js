const { getKegData, getSlackKegData, getSlackSingleKegData } = require('./kegtron');
const axios = require('axios');

class SlackAuth {
    constructor(botToken) {
        this.botToken = botToken;
    }
}

var pendoSlack = new SlackAuth('xoxb-330553142418-1544529695268-HTIcT9Z371ON3OJCSD7KeVS1');

class SlackMessage {
    constructor(reqBody) {
        this._rawBody = { ...reqBody };
        this.user = {
            name: reqBody.user_name || "",
            id: reqBody.user_id || ""
        }
        this.srcCommand = reqBody.command || "";
        this.responseUrl = reqBody.response_url || "";
    }

    sendResponse(data, showInChannel = false, replaceOrig = false) {
        var cfg = {
            method: "post",
            url: this.responseUrl,
            headers: {
                "content-type": "application/json"
            },
            data: data
        }

        if (showInChannel) cfg.data.response_type = "in_channel"
        else cfg.data.response_type = "ephemeral"; // slash commands default to ephemeral
        if (replaceOrig) cfg.data.replace_original = "true";

        return axios(cfg).then((resp) => {
            return true; // true = success
        }).catch((err) => {
            console.log(err);
            console.error("Error sending keg data to Slack response URL")
            return false; // false = failure
        })
    }

    sendTextResponse(message, showInChannel = false, replaceOrig = false) {
        var data = {
            "text": message
        }

        this.sendResponse(data, showInChannel, replaceOrig)
    }

    sendBlockResponse(blockObject, showInChannel = false, replaceOrig = false) {
        var data = blockObject;
        this.sendResponse(data, showInChannel, replaceOrig)
    }
}

class SlackInteractive {
    constructor(reqBody) {
        this._rawBody = {...reqBody};
        this.payload = JSON.parse(reqBody.payload);
        console.log('Slack Interactive Payload: ', this.payload);
        this.user = this.payload.user;
        this.triggerId = this.payload.trigger_id;
        this.actions = this.payload.actions;
        this.responseUrl = this.payload.response_url;
        this.processActions();
    }

    getPostCfg(data) {
        return {
            method: "post",
            url: this.responseUrl,
            headers: {
                "content-type": "application/json"
            },
            data: data
        }
    }

    sendResponse(data, showInChannel = false, replaceOrig = false, deleteOrig = false) {
        var cfg = this.getPostCfg(data);

        if (showInChannel) cfg.data.response_type = "in_channel"
        else cfg.data.response_type = "ephemeral"; // slash commands default to ephemeral
        if (replaceOrig) cfg.data.replace_original = true
        else cfg.data.replace_original = false;
        if (deleteOrig) cfg.data.deleteOrig =  true;

        console.log("Sending cfg: ", cfg);

        return axios(cfg).then((resp) => {
            return true; // true = success
        }).catch((err) => {
            console.log(err);
            console.error("Error sending data to Slack response URL")
            return false; // false = failure
        })
    }

    sendDelete() {
        var cfg = this.getPostCfg({});
        cfg.data.delete_original = "true";
        return axios(cfg).then((resp) => {
            return true; // true = success
        }).catch((err) => {
            console.log(err);
            console.error("Error deleting source Slack message")
            return false; // false = failure
        })
    }

    processActions() {
        this.actions.forEach(action => {
            this.handleAction(action);
        });
    }

    handleAction(action) {
        switch(action.action_id) {
            case "dismiss":
                this.sendDelete();
                break;
            case "share_keg_modal":
                break;
            case "beer_signal_modal":
                break;
            case "share_keg":
                var kegIndex = parseInt(action.value.split('|')[1]);
                Promise.resolve(getSlackSingleKegData(kegIndex, false, this.user.id)).then((data) => {
                    this.sendResponse(data, true)
                });
                break;
            case "beer_signal":
                Promise.resolve(getSlackKegData(false, this.user.id)).then((data) => {
                    this.sendResponse(data, true, null, true);
                });
                break;
        }
    }
}

class SlackModal {
    constructor(triggerId) {
        this.triggerId = triggerId;
        this.view = {};
    }

    trigger() {

    }

    setView() {

    }
}

module.exports = {
    slackMessageHandler: (req, res, next) => {
        var receivedMsg = new SlackMessage(req.body);
        res.status(200).send();
        Promise.resolve(getSlackKegData(true, receivedMsg.user.id)).then((data) => {
            receivedMsg.sendBlockResponse(data);
        });
    },

    slackInteractiveHandler: (req, res, next) => {
        res.status(200).send();
        var receivedInt = new SlackInteractive(req.body);
    }
}