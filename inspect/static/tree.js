import * as vis from "https://unpkg.com/vis-network/standalone/esm/vis-network.min.mjs";

const defaultVisOptions = {
    "physics": {
        "enabled": false
    },
    "layout": {
        "randomSeed": 1,
        "hierarchical": {
            "enabled": true,
            "levelSeparation": 150,
            "nodeSpacing": 200,
            "treeSpacing": 300,
            "blockShifting": false,
            "edgeMinimization": false,
            "parentCentralization": false,
            "direction": "UD",
            "sortMethod": "directed"
        }
    },
    "edges": {
        "smooth": {
            "type": "dynamic",
            "roundness": 1
        },
        "arrows": {
            "to": {
                "enabled": true,
                "scaleFactor": 1
            }
        }
    }
};

function parseOptions(optionsText) {
    try {
        return JSON.parse(optionsText);
    } catch (error) {
        return null;
    }
}

function renderTree(options, nodes, edges) {
    const container = document.getElementById("tree");
    container.innerHTML = ""; // Clear previous render

    const data = {
        nodes: new vis.DataSet(nodes),
        edges: new vis.DataSet(edges),
    };

    try {
        new vis.Network(container, data, options);
    } catch (error) {
        console.error("Error rendering network:", error);
    }
}

function updateOptions(nodes, edges) {
    const textarea = document.getElementById("options-textarea");
    const optionsText = textarea.value.trim();

    if (!optionsText) {
        renderTree(defaultVisOptions, nodes, edges);
        return;
    }

    const parsedOptions = parseOptions(optionsText);
    if (parsedOptions === null) {
        // Invalid JSON - don't render anything
        const container = document.getElementById("tree");
        container.innerHTML = "";
        return;
    }

    renderTree(parsedOptions, nodes, edges);
}

async function fetchTree() {
    return await fetch("/pal/tree.json").then(res => res.json());
}

function sortNodes(nodes) {
    nodes.sort((a, b) => a.inDegree - b.inDegree);
    nodes.sort((a, b) => a.id.localeCompare(b.id));
}

function createNodeTitleTable(node) {
       // Create HTML table for node tooltip using template
       const template = document.getElementById("node-table-template");
       const tableClone = template.content.cloneNode(true);

       const setValue = (selector, value) => tableClone.querySelector(selector).textContent = value;

       // Fill in the table with node properties
       setValue(".node-id", node.id);
       setValue(".node-in-degree", node.inDegree);
       setValue(".node-out-degree", node.outDegree);
       setValue(".node-initer", node.initer);
       setValue(".node-runner", node.runner);
       setValue(".node-health-checker", node.healthChecker);
       setValue(".node-shutdowner", node.shutdowner);

       // Convert the cloned content to HTML string for the title
       const tempDiv = document.createElement('div');
       tempDiv.appendChild(tableClone);
       return tempDiv;
}

function applyNodeStyle(node) {
    node.title = createNodeTitleTable(node);

    if (node.runner) {
        node.shape = "box";
    }

    if (!node.runner && node.inDegree === 0) {
        node.color = "red";
    }

    // make pal components less visible
    if (node.id.includes("github.com/zhulik/pal")) {
        node.opacity = 0.5;
    }

}

(async () => {
    // Initialize textarea with default options
    const textarea = document.getElementById("options-textarea");
    textarea.value = JSON.stringify(defaultVisOptions, null, 2);

    const { nodes, edges } = await fetchTree();

    sortNodes(nodes);

    nodes.forEach(applyNodeStyle);

    // Set up event listener for textarea changes
    textarea.addEventListener('input', () => updateOptions(nodes, edges));

    // Initial render
    renderTree(defaultVisOptions, nodes, edges);
})();