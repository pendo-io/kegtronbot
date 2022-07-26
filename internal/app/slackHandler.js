const { KegTron, KegTronGroup } = require('./kegtron');
const { SlackAuth, SlackAuthGroup, SlackModal, SlackInteractive, SlackMessage } = require('./slackObjects');
const { getSlackAccessTokens, getKegTronDeviceIds } = require('./accessTokens');

async function getAuthControl() {
    if (_authControl) return _authControl;
    else {
        return await refreshAuthControl();
    }
}

function refreshAuthControl() {
    var outObj = new SlackAuthGroup();
    return Promise.resolve(getSlackAccessTokens()).then(slackAuths => {
        slackAuths = slackAuths.data;
        slackAuths.forEach(auth => {
            outObj.addAuth(new SlackAuth(auth.name, auth.bot_token, auth.team_id));
        })
        return outObj;
    });
}

async function getKegTrons() {
    if (_kegTrons) return _kegTrons;
    else {
        return await refreshKegTrons();
    }
}

async function refreshKegTrons() {
    let kegtronGroup = new KegTronGroup();
    let deviceGroups = await getKegTronDeviceIds(); 
    console.warn('deviceGroups: ', deviceGroups)
    for (const deviceGroup of deviceGroups.data) {
        kegtronGroup.addDevice(new KegTron(deviceGroup.device_id_list, deviceGroup.name))
    }
    return kegtronGroup;
}

var _authControl = getAuthControl();
var _kegTrons = getKegTrons();

setInterval(() => {
    // check hourly for updates to auth and device list
    Promise.resolve(getAuthControl()).then(ac => {
        _authControl = ac;
    });
    Promise.resolve(getKegTrons()).then(kt => {
        _kegTrons = kt;
    });
}, 60 * 60 * 1000);

function shareKegMessage(slackInteractive, kegId, customMsg, kegTrons) {
    var deviceName = kegId.split('|')[0];
    var kegIndex = parseInt(kegId.split('|')[1]);
    Promise.resolve(kegTrons.getDevice(deviceName).getSingleKegSlackStatus(kegIndex, false, true, slackInteractive.user.id, customMsg)).then((data) => {
        slackInteractive.sendResponse(data, true)
    });
}

function beerSignalMessage(slackInteractive, deviceName, customMsg, kegTrons) {
    Promise.resolve(kegTrons.getDevice(deviceName).getSlackStatus(false, true, slackInteractive.user.id, customMsg)).then((data) => {
        slackInteractive.sendResponse(data, true, null, true);
    });
}

function processKegActions(slackInteractive, kegTrons) {
    var actions = slackInteractive.getActionsBlock()
    actions.forEach(action => {
        handleKegAction(slackInteractive, action, kegTrons);
    });
}

function handleKegAction(slackInteractive, action, kegTrons) {
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

function handleModalView(slackInteractive, kegTrons) {
    switch (slackInteractive.callbackId) {
        case "share_keg":
            var kegId = slackInteractive.metadata.kegId;
            slackInteractive.setResponseUrl(slackInteractive.payload.response_urls[0].response_url);
            var customMsg = getKegCustomMsg(slackInteractive.stateValues);
            shareKegMessage(slackInteractive, kegId, customMsg, kegTrons);
            break;
        case "beer_signal":
            var deviceName = slackInteractive.metadata.deviceName;
            slackInteractive.setResponseUrl(slackInteractive.payload.response_urls[0].response_url);
            var customMsg = getKegCustomMsg(slackInteractive.stateValues);
            beerSignalMessage(slackInteractive, deviceName, customMsg, kegTrons);
            break;
    }
}

module.exports = {
    slackMessageHandler: async (req, res, next) => {
        res.status(200).send();
        var receivedMsg = new SlackMessage(req.body);
        var ac = await getAuthControl();
        console.log('authControl: ', ac);
        console.log('kegTrons: ', kt);
        if (ac.getBotToken(receivedMsg.getTeamId())) {
            var kt = await getKegTrons();
            // hardcoded to Raleigh right now, but in the future could receive a parameter for different office locations
            Promise.resolve(kt.getDevice('Raleigh').getSlackStatus(true, false, receivedMsg.user.id)).then((data) => {
                receivedMsg.sendResponse(data);
            });
        } else {
            receivedMsg.sendResponse("Workspace not recognized.");
        }
    },

    slackInteractiveHandler: async (req, res, next) => {
        res.status(200).send();
        var ac = await getAuthControl();
        var interactive = new SlackInteractive(req.body, ac);
        if (ac.getBotToken(interactive.getTeamId())) {
            var kt = await getKegTrons();
            if (interactive.isActionBlock()) processKegActions(interactive, kt);
            else if (interactive.isViewSubmit()) handleModalView(interactive, kt);
        } else {
            interactive.sendResponse("Workspace not recognized.");
        }
    }
}
