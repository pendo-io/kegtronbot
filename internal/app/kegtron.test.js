const {KegTron, Keg} = require("./kegtron");

raleighKegtronId = 'S93rEbNyuzVJaDx3sdfaWXQ';

var kegSampleData1 = {
    cData: {
        maker: "Oskar Blue's",
        style: "American Pale Ale",
        userName: " Dale's Pale Ale",
        userDesc: "Beer",
        drinkSize: 355,
    },
    crData: {
        volStart: 19550,
        volDisp: 13724,
    },
    deviceName: "Raleigh",
    portNum: 0
}

var kegSampleData2 = {
    cData: {
        maker: "Sierra Nevada",
        style: "Hazy IPA",
        userName: "Hazy Little Thing",
        userDesc: "Beer",
        drinkSize: 355,
    },
    crData: {
        volStart: 19550,
        volDisp: 9352,
    },
    deviceName: "Raleigh",
    portNum: 1
}

var emptyKegSampleData = {
    cData: {
        maker: "",
        style: "",
        userName: "Empty",
        userDesc: "Kombucha",
        drinkSize: 355,
    },
    crData: {
        volStart: 19550,
        volDisp: 0,
    },
    deviceName: "Raleigh",
    portNum: 1
}

var kegSample1 = new Keg(kegSampleData1.cData, kegSampleData1.crData, kegSampleData1.deviceName, kegSampleData1.portNum);
var kegSample2 = new Keg(kegSampleData2.cData, kegSampleData2.crData, kegSampleData2.deviceName, kegSampleData2.portNum);
var emptyKegSample = new Keg(emptyKegSampleData.cData, emptyKegSampleData.crData, emptyKegSampleData.deviceName, emptyKegSampleData.portNum);

describe('Keg Object', () => {
    test('Keg object is created', () => {
        expect(kegSample1 instanceof Keg).toBe(true);
    });
    
    test('Keg 1 drinks remaining correctly calculated', () => {
        expect(kegSample1.getDrinksRemaining()).toBe(16);
    });
    
    test('Keg 1 is not empty', () => {
        expect(kegSample1.isEmpty()).toBe(false);
    });
    
    test('Empty keg is empty', () => {
        expect(emptyKegSample.isEmpty()).toBe(true);
    })
    
    test('Keg 1 markdown status', () => {
        expect(kegSample1.getMrkdwnStatus()).toBe('*Beer* - _American Pale Ale_ - Oskar Blue\'s | Dale\'s Pale Ale\n `16` drinks remaining')
    })
    
    test('Empty keg markdown status', () => {
        expect(emptyKegSample.getMrkdwnStatus()).toBe('*Kombucha* - Empty\n:x: This keg is empty')
    })
})

describe('KegTron Object', () => {
    test('Live kegtron test', () => {
        var kegTronSample = new KegTron(raleighKegtronId, 'Raleigh');
        Promise.resolve(kegTronSample.updateData()).then(err => {
            expect(err).toEqual(0);
        })
    })
    
    // Remove live updates, replace with fixed sample data
    kegTronSample.updateData = function(){return 0};
    
})
