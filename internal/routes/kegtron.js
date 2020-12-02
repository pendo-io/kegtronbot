const axios = require("axios");
const { slackCompose } = require("./slackCompose");

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
        return Math.floor((this.volumeStart - this.volumeConsumed) / this.drinkSize);
    }

    isEmpty() {
        return this.name.trim().toLowerCase() == "empty" || this.getDrinksRemaining() == 0
    }

    getTextStatus() {
        return `Keg ${this.portnum}: ${this.drinkType} -${this.style ? ` ${this.style} -` : ``}${this.maker ? ` ${this.maker} |` : ``} ${this.name}
${this.isEmpty() ? `This keg is empty` : `${this.getDrinksRemaining()} drinks remaining`}`;
    }

    getMrkdwnStatus(){
        var drRem = this.getDrinksRemaining();
        return `*${this.drinkType}* -${this.style ? ` _${this.style}_ -` : ``}${this.maker ? ` ${this.maker} |` : ``} ${this.name}
${this.isEmpty() ? `:x: This keg is empty` : `${drRem < 10 ? `:warning:` : ``} \`${drRem}\` drinks remaining`}`;
    }

    addSlackStatus(composer, shareBtn = true) {
        var newSection = composer.newSection(this.getMrkdwnStatus(), null);
        if (shareBtn) newSection.setButtonAccessory("Share", `${this.deviceName}|${this.portNum}`, "share_keg");
        composer.addComponent(newSection);
    }
}

class KegTron {
    constructor(deviceId, deviceName) {
        this.rawData = {};
        this.kegs = [];
        this.deviceId = deviceId;
        this.deviceName = deviceName;
        this.lastUpdate = 0;
        this.refresh();
    }

    refresh() {
        var now = Date.now();
        if (now - this.lastUpdate > 60000) {
            return Promise.resolve(this.updateData())
        }
    }

    updateData() {
        var cfg = {
            method: "get",
            url: "https://mdash.net/api/v2/m/device",
            params: {
                access_token: this.deviceId
            }
        }

        return axios(cfg).then((resp) => {
            var data = resp.data.shadow.state.reported;
            this.rawData = data;
            this.buildKegs(data);
            this.lastUpdate = Date.now();
        }).catch((err) => {
            console.log(err);
            console.log('Error retrieving kegtron data');
        })

        
    }

    buildKegs(kegtronData) {
        var i = 0;
        this.kegs = [];
        while (i > -1) {
            var kegPort = `port${i}`;
            if (kegtronData.config[kegPort]) {
                this.kegs.push(new Keg(kegtronData.config[kegPort], kegtronData.config_readonly[kegPort], this.deviceName, i));
                i++;
            } else {
                i = -1;
            }
        }
    }

    getTextStatus() {
        var outMsg = '';
        this.kegs.forEach(keg => {
            outMsg += keg.getTextStatus() + '\n\n';
        })
        return outMsg;
    }

    getSlackStatus(shareBtn, userId) {
        return Promise.resolve(this.refresh()).then(() => {
            var composer = slackCompose();
            this.addKegTronHeader(composer);
            this.addSlackBody(composer, shareBtn, userId);
            if(shareBtn) this.addKegTronActionFooter(composer);
            return composer.json();
        })
    }

    getSlackModal(shareBtn, userId) {
        // Modals triggers are time-sensitive, so we should not refresh before generating the view
        var composer = slackCompose("modal");
        composer.setModalTitle("Share Keg Status");
        composer.setModalSubmitText("Share");
        this.addKegTronHeader(composer);
        this.addSlackBody(composer, shareBtn, userId);
        if(shareBtn) this.addKegTronActionFooter(composer);
        return composer.json();
    }

    addSlackBody(composer, shareBtn, userId) {
        this.kegs.forEach(keg => {
            keg.addSlackStatus(composer, shareBtn);
            composer.addDivider();
        })
        this.addKegTronContext(composer, userId);
    }

    getSingleKegSlackStatus(kegIndex, shareBtn, userId) {
        return Promise.resolve(this.refresh()).then(() => {
            var composer = slackCompose();
            this.addKegTronHeader(composer);
            this.addSingleKegSlackBody(composer, kegIndex, shareBtn, userId);
            return composer.json();
        })
    }

    getSingleKegSlackModal(kegIndex, shareBtn, userId) {
        // Modals triggers are time-sensitive, so we should not refresh before generating the view
        var composer = slackCompose("modal");
        composer.setModalTitle("Share Keg Status");
        composer.setModalSubmitText("Share");
        this.addKegTronHeader(composer);
        this.addSingleKegSlackBody(composer, kegIndex, shareBtn, userId);
        return composer.json();
    }

    addSingleKegSlackBody(composer, kegIndex, shareBtn, userId) {
        this.kegs[kegIndex].addSlackStatus(composer, shareBtn);
        composer.addDivider();
        this.addKegTronContext(composer, userId);
    }

    addKegTronActionFooter(composer) {
        composer.addAction(
            [
                {
                    "type":"button",
                    "text":"Beer Signal",
                    "value":"beer_signal",
                    "action":"beer_signal"
                },
                {
                    "type":"button",
                    "text":"Dismiss",
                    "value":"dismiss",
                    "action":"dismiss"
                }
            ]
        );
    }

    addKegTronHeader(composer) {
        composer.addHeader(":beers: Check Out What's On Tap! :beers:")
        composer.addDivider();
    } 

    addKegTronContext(composer, userId) {
        composer.addContext(`:beer: Shared${userId ? ` by <@${userId}>` : ``} via */kegtron*`);
    } 
}

var kegTronRaleigh = new KegTron('S93rEbNyuzVJaDx3sdfaWXQ', 'Raleigh');

module.exports = {
    getKegData: () => {
        return kegTronRaleigh.getTextStatus();
    },

    getSlackKegData: (shareBtn, userId) => {
        return Promise.resolve(kegTronRaleigh.getSlackStatus(shareBtn, userId));
    },

    getSlackSingleKegData: (kegIndex, shareBtn, userId) => {
        return Promise.resolve(kegTronRaleigh.getSingleKegSlackStatus(kegIndex, shareBtn, userId));
    }
}
