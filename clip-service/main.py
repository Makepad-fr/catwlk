import socket
import json
import base64
from PIL import Image
import io

import torch
import clip

# --- Only these dependencies: torch, clip, pillow ---

device = "cpu"
model, preprocess = clip.load("ViT-B/32", device=device)

def classify(image_b64, labels_dict):
    # Decode image
    image_bytes = base64.b64decode(image_b64)
    image = Image.open(io.BytesIO(image_bytes)).convert("RGB")
    image_input = preprocess(image).unsqueeze(0).to(device)

    results = {}
    for category, labels in labels_dict.items():
        text_inputs = torch.cat([clip.tokenize(label) for label in labels]).to(device)
        with torch.no_grad():
            image_features = model.encode_image(image_input)
            text_features = model.encode_text(text_inputs)
            logits_per_image, _ = model(image_input, text_inputs)
            probs = logits_per_image.softmax(dim=-1).cpu().numpy()[0]
        idx = probs.argmax()
        results[category] = {
            "label": labels[idx],
            "score": float(probs[idx])
        }
    return results

def handle_rpc(request_json):
    req = json.loads(request_json)
    method = req["method"]
    params = req["params"]
    rpc_id = req["id"]

    if method == "classify":
        try:
            result = classify(params["image"], params["labels"])
            response = {"jsonrpc": "2.0", "result": result, "id": rpc_id}
        except Exception as e:
            response = {
                "jsonrpc": "2.0",
                "error": {"code": -32000, "message": str(e)},
                "id": rpc_id
            }
    else:
        response = {
            "jsonrpc": "2.0",
            "error": {"code": -32601, "message": "Method not found"},
            "id": rpc_id
        }
    return json.dumps(response)

def start_server(host="0.0.0.0", port=9000):
    import socket

    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind((host, port))
    server.listen(1)
    print(f"TCP JSON-RPC server listening on {host}:{port}")

    while True:
        conn, addr = server.accept()
        with conn:
            buffer = b""
            # Read until a newline (\n) is found
            while True:
                chunk = conn.recv(4096)
                if not chunk:
                    break
                buffer += chunk
                if b"\n" in chunk:
                    break
            req_str = buffer.decode().strip()
            if not req_str:
                continue
            print("Received:", req_str[:120], "..." if len(req_str) > 120 else "")
            resp_str = handle_rpc(req_str)
            print("Sending:", resp_str[:120], "..." if len(resp_str) > 120 else "")
            conn.sendall((resp_str + "\n").encode())  # Send response, newline-terminated



if __name__ == "__main__":
    start_server()
