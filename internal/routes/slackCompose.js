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

    addTextInput(text, actionId, blockId, isOpt) {
        this.addComponent(new SlackBlockTextInput(text, actionId, blockId, isOpt));
    }
    
    addChannelSelect(text, actionId, blockId, isOpt) {
        this.addComponent(new SlackBlockChannelSelect(text, actionId, blockId, isOpt));
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
        console.log("Type is: ", this.type)
        if (this.type == "modal") {
            console.log("Adding modal fluff...");
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
    constructor(text, actionId, blockId, isOptional = false) {
        super();
        this.text = text || "";
        if (actionId) this.actionId = actionId;
        if (blockId) this.blockId = blockId;
        this.isOptional = !!isOptional;
    }

    json() {
        var outObj = {
            "type": "input",
            "optional": this.isOptional,
			"element": {
				"type": "plain_text_input"
			},
			"label": {
				"type": "plain_text",
				"text": this.text,
				"emoji": true
			}
        }
        if (this.blockId) outObj.block_id = this.blockId;
        if (this.actionId) outObj.element.action_id = this.actionId;
        return outObj;
    }
}

class SlackBlockChannelSelect extends SlackBlockComponent {
    constructor(text, actionId, blockId, isOptional = false) {
        super();
        this.text = text || "";
        if (actionId) this.actionId = actionId;
        if (blockId) this.blockId = blockId;
        this.isOptional = !!isOptional;
    }

    json() {
        var outObj = {
            "type": "input",
			"optional": this.isOptional,
			"label": {
				"type": "plain_text",
				"text": this.text
			},
			"element": {
				"default_to_current_conversation": true,
				"type": "conversations_select",
				"response_url_enabled": true
			}
        }
        if (this.blockId) outObj.block_id = this.blockId;
        if (this.actionId) outObj.element.action_id = this.actionId;
        return outObj;
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
    slackCompose: (type) => {
        return new SlackCompose(type);
    }
}