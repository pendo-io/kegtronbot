const { Storage } = require('@google-cloud/storage');
const axios = require('axios');

// Instantiate a storage client
const storage = new Storage();
const bucket = "access_tokens";

async function generateV4ReadSignedUrl(bucketName, fileName) { //https://cloud.google.com/storage/docs/access-control/signing-urls-with-helpers#code-samples
    // These options will allow temporary read access to the file
    const options = {
        version: 'v4',
        action: 'read',
        expires: Date.now() + 3 * 60 * 1000, // 3 minutes
    };

    // Get a v4 signed URL for reading the file
    // Note: service account must be given "Service Account Token Creator" role to run this operation
    // https://cloud.google.com/iam/docs/service-accounts#token-creator-role
    const [url] = await storage.bucket(bucketName).file(fileName).getSignedUrl(options);
    return url;
}

module.exports = {
    getSlackAccessTokens: () => {
        return Promise.resolve(generateV4ReadSignedUrl(bucket, "slackauths.json").catch(console.error)).then(url => {
            return axios.get(url).then(response => {
                return response.data;
            });
        });
    },

    getKegTronDeviceIds: () => {
        return Promise.resolve(generateV4ReadSignedUrl(bucket, "kegtrons.json").catch(console.error)).then(url => {
            return axios.get(url).then(response => {
                return response.data;
            });
        });
    }
}