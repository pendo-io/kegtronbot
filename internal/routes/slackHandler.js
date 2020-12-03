const { KegTron } = require('./kegtron');
const {SlackAuth, SlackAuthGroup, SlackModal, SlackInteractive, SlackMessage} = require('./slackObjects');

var authControl = new SlackAuthGroup();
authControl.addAuth(new SlackAuth('pendo-test', 'xoxb-330553142418-1544529695268-HTIcT9Z371ON3OJCSD7KeVS1', 'T9QG946CA'));

var kegTronRaleigh = new KegTron('S93rEbNyuzVJaDx3sdfaWXQ', 'Raleigh');

function shareKegMessage(slackInteractive, kegId, customMsg) {
    var kegIndex = parseInt(kegId.split('|')[1]);
    Promise.resolve(kegTronRaleigh.getSingleKegSlackStatus(kegIndex, false, true, slackInteractive.user.id, customMsg)).then((data) => {
        slackInteractive.sendResponse(data, true)
    });
}

function beerSignalMessage(slackInteractive, customMsg) {
    Promise.resolve(kegTronRaleigh.getSlackStatus(false, true, slackInteractive.user.id, customMsg)).then((data) => {
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
            var kegIndex = parseInt(action.value.split('|')[1]);
            var modalView = kegTronRaleigh.getSingleKegSlackModal(kegIndex, false, slackInteractive.user.id);
            var shareModal = new SlackModal(slackInteractive.triggerId, modalView, slackInteractive.botToken);
            shareModal.trigger('share_keg', JSON.stringify({ 'kegId': action.value }));
            break;
        case "beer_signal_modal":
            var modalView = kegTronRaleigh.getSlackModal(false, slackInteractive.user.id);
            var shareModal = new SlackModal(slackInteractive.triggerId, modalView, slackInteractive.botToken);
            shareModal.trigger('beer_signal');
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
            slackInteractive.setResponseUrl(slackInteractive.payload.response_urls[0].response_url);
            var customMsg = getKegCustomMsg(slackInteractive.stateValues);
            beerSignalMessage(slackInteractive, customMsg);
            break;
    }
}

module.exports = {
    slackMessageHandler: (req, res, next) => {
        var receivedMsg = new SlackMessage(req.body);
        if (authControl.getBotToken(receivedMsg.getTeamId())) {
            res.status(200).send();
            Promise.resolve(kegTronRaleigh.getSlackStatus(true, false, receivedMsg.user.id)).then((data) => {
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