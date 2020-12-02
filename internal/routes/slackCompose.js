class SlackCompose {
    constructor(type) {
        this.type = type;
        this.components = [];
    }

    addComponent(comp) {
        if (comp instanceof SlackBlockComponent) this.components.push(comp);
        else {
            console.error("Error - tried to add non-block component");
            console.log("Component: ", comp);
        }
    }

    addDivider() {
        this.addComponent(new SlackBlockDivider());
    }

    addHeader(text) {
        this.addComponent(new SlackBlockHeader(text));
    }

    addContext(text) {
        this.addComponent(new SlackBlockContext(text));
    }

    addSection(text, accessory) {
        this.addComponent(new SlackBlockSection(text, accessory));
    }

    addAction(actionArr) {
        this.addComponent(new SlackBlockAction(actionArr));
    }

    newSection(text, accessory) {
        return new SlackBlockSection(text, accessory);
    }

    newAction() {
        return new SlackBlockAction(actionArr);
    }

    delComponent(compIndex) {
        delete this.components[compIndex];
    }

    setModalTitle(titleText) {
        this.titleText = titleText;
    }

    setModalSubmitText(submitText) {
        this.submitText = submitText;
    }

    json() {
        var outObj = {};
        outObj.blocks = [];
        this.components.forEach(component => {
            outObj.blocks.push(component.json());
        })
        if (this.type == "modal") {
            outObj.type = "modal";
            outObj.title = {
                "type": "plain_text",
                "text": this.titleText || "",
                "emoji": true
            },
            outObj.submit = {
                "type": "plain_text",
                "text": this.submitText || "Submit",
                "emoji": true
            }
        }
        return outObj;

    }
}

class SlackBlockComponent {
    json() {
        return {};
    }
}

class SlackBlockDivider extends SlackBlockComponent {
    constructor() {
        super();
    }

    json() {
        return { "type": "divider" };
    }
}

class SlackBlockHeader extends SlackBlockComponent {
    constructor(text) {
        super();
        this.text = text;
    }

    json() {
        return { 
            "type": "header",
			"text": {
				"type": "plain_text",
				"text": this.text,
				"emoji": true
			}
         };
    }
}

class SlackBlockContext extends SlackBlockComponent {
    constructor(text) {
        super();
        this.text = text;
    }

    json() {
        return {
            "type": "context",
			"elements": [
				{
					"type": "mrkdwn",
					"text": this.text
				}
			]
        }
    }
}

class SlackBlockSection extends SlackBlockComponent {
    constructor(text, accessory) {
        super();
        this.text = text || "";
        this.accessory = accessory;
    }

    setButtonAccessory(text, value, action) {
        this.accessory = new SlackBlockAccessoryButton(text, value, action);
    }

    json() {
        var outObj = {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": this.text,
            },
        }
        if (this.accessory) outObj.accessory = this.accessory.json();
        return outObj;
    }
}

class SlackBlockTextInput extends SlackBlockComponent {
    constructor(text, actionId) {
        super();
        this.text = text || "";
        this.actionId = actionId || "";
    }

    json() {
        return {
            "type": "input",
			"element": {
				"type": "plain_text_input",
				"action_id": this.actionId
			},
			"label": {
				"type": "plain_text",
				"text": this.text,
				"emoji": true
			}
        }
    }
}

class SlackBlockAction extends SlackBlockComponent {
    constructor(actionArr) {
        super();
        this.actions = [];
        actionArr.forEach(action => {
            this.addAction(action);
        })
    }

    addAction(props) {
        switch(props.type) {
            case "button":
                this.addButton(props.text, props.value, props.action);
                break;
        }
    }

    addButton(text, value, action) {
        this.actions.push(new SlackBlockAccessoryButton(text, value, action));
    }

    json() {
        var outObj = {
            "type": "actions",
            "elements": []
        }
        this.actions.forEach(action => {
            outObj.elements.push(action.json());
        })
        return outObj;
    }
}

class SlackBlockAccessory {
    json() {
        return {}
    }
}

class SlackBlockAccessoryButton extends SlackBlockAccessory {
    constructor(text, value, action) {
        super();
        this.text = text;
        this.value = `${value}`; // must be a string
        this.action = action;
    }

    json() {
        return {
            "type": "button",
            "text": {
                "type": "plain_text",
                "text": this.text,
                "emoji": true
            },
            "value": this.value,
            "action_id": this.action,
        }
    }
}

module.exports = {
    slackCompose: () => {
        return new SlackCompose();
    }
}