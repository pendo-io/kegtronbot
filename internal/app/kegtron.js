const axios = require("axios");
const { SlackCompose } = require("./slackCompose");

class Keg {
    constructor(cData, crData, deviceName, portNum) {
        this.maker = cData.maker;
        this.style = cData.style;
        this.name = cData.userName.trim();
        this.drinkType = cData.userDesc;
        this.drinkSize = cData.drinkSize;
        this.volumeStart = crData.volStart;
        this.volumeConsumed = crData.volDisp;
        this.deviceName = deviceName;
        this.portNum = portNum;
    }

    getDrinksRemaining() {
        return (((this.volumeStart - this.volumeConsumed) / this.volumeStart) * 100).toFixed(1);
    }

    isEmpty() {
        return this.name.trim().toLowerCase() == "empty" || this.getDrinksRemaining() < 0.2
    }

    getTextStatus() {
        return `Keg ${this.portnum}: ${this.drinkType} -${this.style ? ` ${this.style} -` : ``}${this.maker ? ` ${this.maker} |` : ``} ${this.name}
${this.isEmpty() ? `This keg is empty` : `${this.getDrinksRemaining()}% remaining`}`;
    }

    getMrkdwnStatus() {
        var drRem = this.getDrinksRemaining();
        return `*${this.name}* (${this.maker})
_${this.style}_
${this.isEmpty() ? `:x: This keg is empty` : `${drRem < 5 ? `:warning:` : ``} \`${drRem}% remaining\``}`;
    }

    addSlackStatus(composer, shareBtn = true) {
        var newSection = composer.newSection(this.getMrkdwnStatus(), null);
        if (shareBtn) newSection.setButtonAccessory("Share", `${this.deviceName}|${this.portNum}`, "share_keg_modal");
        composer.addComponent(newSection);
    }
}

class KegTron {
    constructor(deviceIdList, deviceName) {
        this.rawData = {};
        this.kegs = [];
        this.deviceIdList = deviceIdList;
        this.deviceName = deviceName;
        this.lastUpdate = 0;
        this.refresh();
    }

    getName() {
        return this.deviceName;
    }

    refresh() {
        var now = Date.now();
        if (now - this.lastUpdate > 60000) {
            return Promise.resolve(this.updateData())
        }
    }

    async updateData() {
        let rawData = [];

        try {
            var requestList = this.deviceIdList.map(async (id) => {
                let config = {
                    method: "get",
                    url: "https://mdash.net/api/v2/m/device",
                    params: {
                        access_token: id
                    }
                }

                let resp = await axios(config);
                let data = resp.data.shadow.state.reported;
                rawData.push(data);
            });
        } catch (err) {
            console.error(err);
            console.error('Error retrieving kegtron data');
            return 1;
        }
        await Promise.all(requestList);
        this.rawData = rawData;

        this.buildKegs(rawData);
        this.lastUpdate = Date.now();
        return 0;
    }

    buildKegs(kegtronData) {
        let newKegs = []
        for (const kegtronDevice of kegtronData) {
            var i = 0;
            while (i > -1) {
                var kegPort = `port${i}`;
                if (kegtronDevice.config[kegPort]) {
                    newKegs.push(new Keg(kegtronDevice.config[kegPort], kegtronDevice.config_readonly[kegPort], this.deviceName, i));
                    i++;
                } else {
                    i = -1;
                }
            }
        }
        this.kegs = newKegs;
    }

    getTextStatus() {
        var outMsg = '';
        this.kegs.forEach(keg => {
            outMsg += keg.getTextStatus() + '\n\n';
        })
        return outMsg;
    }

    getSlackStatus(shareBtn, showContext, userId, customMsg) {
        return Promise.resolve(this.refresh()).then(() => {
            var composer = new SlackCompose();
            this.addKegTronHeader(composer);
            if (customMsg) this.addKegTronCustomMessage(composer, customMsg);
            this.addSlackBody(composer, shareBtn, userId);
            if (showContext) this.addKegTronContext(composer, userId);
            if (shareBtn) this.addKegTronActionFooter(composer);
            return composer.json();
        })
    }

    getSlackModal(shareBtn, userId) {
        // Modals triggers are time-sensitive, so we should not refresh before generating the view
        var composer = new SlackCompose("modal");
        composer.setModalTitle("Share Keg Status");
        composer.setModalSubmitText("Share");
        this.addKegTronChannelSelect(composer);
        //this.addKegTronHeader(composer);
        this.addKegTronCustomMessageInput((composer));
        this.addSlackBody(composer, shareBtn, userId);
        if (shareBtn) this.addKegTronActionFooter(composer);
        return composer.json();
    }

    addSlackBody(composer, shareBtn, userId) {
        if (shareBtn) {
            this.kegs.forEach(keg => {
                keg.addSlackStatus(composer, shareBtn);
            })
        } else {
            for (var i = 0; i < this.kegs.length; i=i+2) {
                var keg1 = this.kegs[i];
                var keg2 = this.kegs[i+1];
                var newSection = composer.newSection(keg1.getMrkdwnStatus(), null);
                if (keg2) newSection.addTextBlock(keg2.getMrkdwnStatus())
                else newSection.addTextBlock(" ");
                composer.addComponent(newSection);
            }
        }
        composer.addDivider();
    }

    getSingleKegSlackStatus(kegIndex, shareBtn, showContext, userId, customMsg) {
        return Promise.resolve(this.refresh()).then(() => {
            var composer = new SlackCompose();
            this.addKegTronHeader(composer);
            if (customMsg) this.addKegTronCustomMessage(composer, customMsg);
            this.addSingleKegSlackBody(composer, kegIndex, shareBtn, userId);
            if (showContext) this.addKegTronContext(composer, userId);
            return composer.json();
        })
    }

    getSingleKegSlackModal(kegIndex, shareBtn, userId) {
        // Modals triggers are time-sensitive, so we should not refresh before generating the view
        var composer = new SlackCompose("modal");
        composer.setModalTitle("Share Keg Status");
        composer.setModalSubmitText("Share");
        this.addKegTronChannelSelect(composer);
        //this.addKegTronHeader(composer);
        this.addKegTronCustomMessageInput((composer));
        this.addSingleKegSlackBody(composer, kegIndex, shareBtn, userId);
        return composer.json();
    }

    addSingleKegSlackBody(composer, kegIndex, shareBtn, userId) {
        this.kegs[kegIndex].addSlackStatus(composer, shareBtn);
        composer.addDivider();
    }

    addKegTronActionFooter(composer) {
        composer.addAction(
            [
                {
                    "type": "button",
                    "text": "Beer Signal",
                    "value": this.deviceName,
                    "action": "beer_signal_modal"
                },
                {
                    "type": "button",
                    "text": "Dismiss",
                    "value": "dismiss",
                    "action": "dismiss"
                }
            ]
        );
    }

    addKegTronHeader(composer) {
        composer.addHeader(":beers: Check Out What's On Tap! :beers:")
        composer.addDivider();
    }

    addKegTronCustomMessageInput(composer) {
        composer.addTextInput("Custom Message", "custom_message_input","custom_message_block", true);
        composer.addDivider();
    }

    addKegTronCustomMessage(composer, text) {
        composer.addSection(text);
        composer.addDivider();
    }

    addKegTronChannelSelect(composer) {
        composer.addChannelSelect("Choose a Channel to Share In", "channel_select","channel_select_block", false);
        composer.addDivider();
    }

    addKegTronContext(composer, userId) {
        composer.addContext(`:beer: Shared${userId ? ` by <@${userId}>` : ``} via */kegtron*`);
    }
}

class KegTronGroup {
    constructor() {
        this.devices = {};
    }

    addDevice(newDevice) {
        if (newDevice instanceof KegTron) this.devices[newDevice.getName()] = newDevice;
        else {
            console.error("Error -- attempted to add object that is not a valid KegTron device");
            console.error("Added object: ", newDevice);
        }
    }

    getDevice(name) {
        return this.devices[name] || {};
    }
}

module.exports = {
    Keg,
    KegTron,
    KegTronGroup
}
