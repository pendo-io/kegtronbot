const axios = require('axios');
class SlackAuth {
    constructor(name, botToken, teamId) {
        this.name = name;
        this._botToken = botToken;
        this._teamId = teamId;
    }

    getBotToken() {
        return this._botToken;
    }

    checkTeamId(teamIdCheck) {
        return this._teamId == teamIdCheck;
    }
}

class SlackAuthGroup {
    constructor() {
        this.workspaces = [];
    }

    addAuth(newAuth) {
        if (newAuth instanceof SlackAuth) this.workspaces.push(newAuth);
        else {
            console.error("Error -- attempted to add object that is not a valid SlackAuth");
            console.error("Added object: ", newAuth);
        }
    }

    getBotToken(teamId) {
        var outToken = "";
        this.workspaces.forEach(team => {
            console.log(`Checking if team ${teamId} matches ${team._teamId}. Result: ${team.checkTeamId(teamId)}`);
            if (team.checkTeamId(teamId)) outToken = team.getBotToken();
        })
        return outToken;
    }
}

class SlackMessage {
    constructor(reqBody) {
        this._rawBody = { ...reqBody };
        console.log("Message Body: ", reqBody);
        this.user = {
            name: reqBody.user_name || "",
            id: reqBody.user_id || ""
        }
        this.srcCommand = reqBody.command || "";
        this.responseUrl = reqBody.response_url || "";
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

    sendResponse(data, showInChannel = false, replaceOrig = false) {
        var cfg = this.getPostCfg(data);

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
}

class SlackInteractive {
    constructor(reqBody, slackAuthGroup) {
        this._rawBody = { ...reqBody };
        this.payload = JSON.parse(reqBody.payload);
        this.type = this.payload.type;
        console.log('Slack Interactive Payload: ', this.payload);
        this.user = this.payload.user;
        this.team = this.payload.team;
        console.log('Team ID: ', this.team.id);
        this.botToken = slackAuthGroup.getBotToken(this.team.id);
        console.log('Returned token: ', this.botToken);
        this.triggerId = this.payload.trigger_id;
        this.actions = this.payload.actions;
        switch (this.type) {
            case "block_actions":
                this.stateValues = "";
                this.callbackId = "";
                this.metadata = "";
                this.responseUrl = this.payload.response_url;
                this.actions = this.payload.actions;
                break;
            case "view_submission":
                this.stateValues = this.payload.view.state.values;
                this.callbackId = this.payload.view.callback_id;
                this.metadata = (this.payload.view.private_metadata ? JSON.parse(this.payload.view.private_metadata) : {});
                this.responseUrl = "";
                this.actions = [];
                break;
        }
    }

    isActionBlock() {
        return this.type == "block_actions";
    }

    getActionsBlock() {
        return this.actions;
    }

    isViewSubmit() {
        return this.type == "view_submission";
    }

    getCallbackId() {
        return this.callbackId;
    }

    getStateValues() {
        return this.stateValues;
    }

    setResponseUrl(respUrl) {
        this.responseUrl = respUrl;
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
        if (deleteOrig) cfg.data.deleteOrig = true;

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
}

class SlackModal {
    constructor(triggerId, modalView, botToken) {
        this.triggerId = triggerId;
        this.view = modalView;
        this.botToken = botToken
    }

    getPostCfg(data) {
        return {
            method: "post",
            headers: {
                "content-type": "application/json",
                Authorization: `Bearer ${this.botToken}`
            },
            data: data
        }
    }

    trigger(callbackId, metadata) {
        var data = {};
        data.view = this.view;
        data.view.private_metadata = metadata || "";
        data.view.callback_id = callbackId;
        data.trigger_id = this.triggerId;
        var cfg = this.getPostCfg(data);
        cfg.url = "https://slack.com/api/views.open";
        axios(cfg).then((resp) => {
            resp = resp.data;
        }).catch(err => {
            console.log(err);
            console.log("Error triggering modal");
        })
    }

    setView(modalView) {
        this.view = modalView;
    }
}

module.exports = {
    SlackAuth,
    SlackAuthGroup,
    SlackMessage,
    SlackInteractive,
    SlackModal
}