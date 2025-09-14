import * as vis from "https://unpkg.com/vis-network/standalone/esm/vis-network.min.mjs";

const defaultVisOptions = {
    "physics": {
        "enabled": true,
    },
    "layout": {
        "randomSeed": 1,
        "hierarchical": {
            "levelSeparation": 80
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

(async () => {
    const { nodes, edges } = await fetch("/pal/tree.json").then(res => res.json());

    nodes.sort((a, b) => a.inDegree - b.inDegree);
    nodes.sort((a, b) => a.id.localeCompare(b.id));

    nodes.forEach(node => {
        if (node.runner) {
            node.shape = "box";
        }
    });

    // Initialize textarea with default options
    const textarea = document.getElementById("options-textarea");
    textarea.value = JSON.stringify(defaultVisOptions, null, 2);

    // Set up event listener for textarea changes
    textarea.addEventListener('input', () => updateOptions(nodes, edges));

    // Initial render
    renderTree(defaultVisOptions, nodes, edges);
})();