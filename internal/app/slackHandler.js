const { KegTron, KegTronGroup } = require('./kegtron');
const {SlackAuth, SlackAuthGroup, SlackModal, SlackInteractive, SlackMessage} = require('./slackObjects');
const {getSlackAccessTokens, getKegTronDeviceIds} = require('./accessTokens');


function getAuthControl() {
    var outObj = new SlackAuthGroup();
    return Promise.resolve(getSlackAccessTokens()).then(slackAuths => {
        slackAuths = slackAuths.data;
        slackAuths.forEach(auth => {
            outObj.addAuth(new SlackAuth(auth.name, auth.bot_token, auth.team_id));
        })
        return outObj;
    });
}

function getKegTrons() {
    var outObj = new KegTronGroup();
    return Promise.resolve(getKegTronDeviceIds()).then( devices => {
        devices = devices.data;
        devices.forEach(device => {
            outObj.addDevice(new KegTron(device.device_id, device.name));
        })
        return outObj;
    });
}

var authControl, kegTrons;
Promise.resolve(getAuthControl()).then(ac => {
    authControl = ac;
});
Promise.resolve(getKegTrons()).then(kt => {
    kegTrons = kt;
});

setInterval(() => {
    // check hourly for updates to auth and device list
    Promise.resolve(getAuthControl()).then(ac => {
        authControl = ac;
    });
    Promise.resolve(getKegTrons()).then(kt => {
        kegTrons = kt;
    });
}, 60 * 60 * 1000);

function shareKegMessage(slackInteractive, kegId, customMsg) {
    var deviceName = kegId.split('|')[0];
    var kegIndex = parseInt(kegId.split('|')[1]);
    Promise.resolve(kegTrons.getDevice(deviceName).getSingleKegSlackStatus(kegIndex, false, true, slackInteractive.user.id, customMsg)).then((data) => {
        slackInteractive.sendResponse(data, true)
    });
}

function beerSignalMessage(slackInteractive, deviceName, customMsg) {
    Promise.resolve(kegTrons.getDevice('Raleigh').getSlackStatus(false, true, slackInteractive.user.id, customMsg)).then((data) => {
        slackInteractive.sendResponse(data, true, null, true);
    });
}

function processKegActions(slackInteractive) {
    var actions = slackInteractive.getActionsBlock()
    actions.forEach(action => {
        handleKegAction(slackInteractive, action);
    });
}

function handleKegAction(slackInteractive, action) {
    switch (action.action_id) {
        case "dismiss":
            slackInteractive.sendDelete();
            break;
        case "share_keg_modal":
            var deviceName = action.value.split('|')[0]
            var kegIndex = parseInt(action.value.split('|')[1]);
            var modalView = kegTrons.getDevice(deviceName).getSingleKegSlackModal(kegIndex, false, slackInteractive.user.id);
            var shareModal = new SlackModal(slackInteractive.triggerId, modalView, slackInteractive.botToken);
            shareModal.trigger('share_keg', JSON.stringify({ 'kegId': action.value }));
            break;
        case "beer_signal_modal":
            var deviceName = action.value;
            var modalView = kegTrons.getDevice(deviceName).getSlackModal(false, slackInteractive.user.id);
            var shareModal = new SlackModal(slackInteractive.triggerId, modalView, slackInteractive.botToken);
            shareModal.trigger('beer_signal', JSON.stringify({ 'deviceName': action.value }));
            break;
    }
}

function getKegCustomMsg(stateValues) {
    return stateValues.custom_message_block.custom_message_input.value;
}

function handleModalView(slackInteractive) {
    switch (slackInteractive.callbackId) {
        case "share_keg":
            var kegId = slackInteractive.metadata.kegId;
            slackInteractive.setResponseUrl(slackInteractive.payload.response_urls[0].response_url);
            var customMsg = getKegCustomMsg(slackInteractive.stateValues);
            shareKegMessage(slackInteractive, kegId, customMsg);
            break;
        case "beer_signal":
            var deviceName = slackInteractive.metadata.deviceName;
            slackInteractive.setResponseUrl(slackInteractive.payload.response_urls[0].response_url);
            var customMsg = getKegCustomMsg(slackInteractive.stateValues);
            beerSignalMessage(slackInteractive, deviceName, customMsg);
            break;
    }
}

module.exports = {
    slackMessageHandler: (req, res, next) => {
        var receivedMsg = new SlackMessage(req.body);
        console.log('authControl: ', authControl);
        console.log('kegTrons: ', kegTrons);
        if (authControl.getBotToken(receivedMsg.getTeamId())) {
            res.status(200).send();
            Promise.resolve(kegTrons.getDevice('Raleigh').getSlackStatus(true, false, receivedMsg.user.id)).then((data) => {
                receivedMsg.sendResponse(data);
            });
        } else {
            res.status(403).type('txt').send("Workspace not recognized.")
        }
    },

    slackInteractiveHandler: (req, res, next) => {
        var interactive = new SlackInteractive(req.body, authControl);
        if (authControl.getBotToken(interactive.getTeamId())) {
            res.status(200).send();
            if (interactive.isActionBlock()) processKegActions(interactive);
            else if (interactive.isViewSubmit()) handleModalView(interactive);
        } else {
            res.status(403).type('txt').send("Workspace not recognized.")
        }
    }
}