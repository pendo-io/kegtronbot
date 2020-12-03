const { getKegData, getSlackKegData, getSlackSingleKegData, getSlackKegModal, getSlackSingleKegModal} = require('./kegtron');
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
        this.type = this.payload.type;
        console.log('Slack Interactive Payload: ', this.payload);
        this.user = this.payload.user;
        this.triggerId = this.payload.trigger_id;
        this.actions = this.payload.actions;
        switch (this.type) {
            case "block_actions":
                this.responseUrl = this.payload.response_url;
                this.actions = this.payload.actions;
                this.processActions();
                break;
            case "view_submission":
                this.stateValues = this.payload.view.state.values;
                this.callbackId = this.payload.view.callback_id;
                this.metadata = (this.payload.view.private_metadata ? JSON.parse(this.payload.view.private_metadata) : {});
                console.log('Metadata: ', this.metadata);
                console.log('State: ', this.stateValues);
                this.handleModalView();
                break;
        }
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

    shareKegMessage(kegId, customMsg) {
        var kegIndex = parseInt(kegId.split('|')[1]);
        Promise.resolve(getSlackSingleKegData(kegIndex, false, true, this.user.id, customMsg)).then((data) => {
            this.sendResponse(data, true)
        });
    }

    beerSignalMessage(customMsg) {
        Promise.resolve(getSlackKegData(false, true, this.user.id, customMsg)).then((data) => {
            this.sendResponse(data, true, null, true);
        });
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
                var kegIndex = parseInt(action.value.split('|')[1]);
                var modalView = getSlackSingleKegModal(kegIndex, false, this.user.id);
                var shareModal = new SlackModal(this.triggerId, modalView);
                shareModal.trigger('share_keg', JSON.stringify({'kegId':action.value}));
                break;
            case "beer_signal_modal":
                var modalView = getSlackKegModal(false, this.user.id);
                var shareModal = new SlackModal(this.triggerId, modalView);
                shareModal.trigger('beer_signal');
                break;
            case "share_keg":
                this.shareKegMessage(action.value);
                break;
            case "beer_signal":
                this.beerSignalMessage();
                break;
        }
    }

    getCustomMsg(stateValues) {
        return stateValues.custom_message_block.custom_message_input.value;
    }

    handleModalView() {
        switch(this.callbackId) {
            case "share_keg":
                var kegId = this.metadata.kegId;
                this.responseUrl = this.payload.response_urls[0].response_url;
                var customMsg = this.getCustomMsg(this.stateValues);
                this.shareKegMessage(kegId, customMsg);
                break;
            case "beer_signal":
                this.responseUrl = this.payload.response_urls[0].response_url;
                var customMsg = this.getCustomMsg(this.stateValues);
                this.beerSignalMessage(customMsg);
                break;
        }
    }
}

class SlackModal {
    constructor(triggerId, modalView) {
        this.triggerId = triggerId;
        this.view = modalView;
    }

    getPostCfg(data) {
        return {
            method: "post",
            headers: {
                "content-type": "application/json",
                Authorization: `Bearer ${pendoSlack.botToken}`
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
        console.log('Sending data from trigger: ', data);
        cfg.url = "https://slack.com/api/views.open";
        axios(cfg).then((resp) => {
            resp = resp.data;
            console.log('Sent Modal Post. Response Data:');
            console.log(resp);
        }).catch(err =>{
            console.log(err);
            console.log("Error triggering modal");
        })
    }

    setView(modalView) {
        this.view = modalView;
    }
}

module.exports = {
    slackMessageHandler: (req, res, next) => {
        var receivedMsg = new SlackMessage(req.body);
        res.status(200).send();
        Promise.resolve(getSlackKegData(true, false, receivedMsg.user.id)).then((data) => {
            receivedMsg.sendBlockResponse(data);
        });
    },

    slackInteractiveHandler: (req, res, next) => {
        res.status(200).send();
        var receivedInt = new SlackInteractive(req.body);
    }
}