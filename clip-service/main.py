import socket
import json

def add(a, b):
    return a + b

def handle_rpc(request_json):
    req = json.loads(request_json)
    method = req["method"]
    params = req["params"]
    rpc_id = req["id"]

    # Only 'add' supported in this example
    if method == "add":
        result = add(params["a"], params["b"])
        response = {"jsonrpc": "2.0", "result": result, "id": rpc_id}
    else:
        response = {
            "jsonrpc": "2.0",
            "error": {"code": -32601, "message": "Method not found"},
            "id": rpc_id
        }
    return json.dumps(response)

def start_server(host="127.0.0.1", port=9000):
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind((host, port))
    server.listen(1)
    print(f"TCP JSON-RPC server listening on {host}:{port}")

    while True:
        conn, addr = server.accept()
        with conn:
            data = conn.recv(4096)
            if not data:
                continue
            req_str = data.decode()
            print("Received:", req_str)
            resp_str = handle_rpc(req_str)
            print("Sending:", resp_str)
            conn.sendall(resp_str.encode())

if __name__ == "__main__":
    start_server()
