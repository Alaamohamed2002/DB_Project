from flask import Flask, request, render_template_string
import socket

app = Flask(__name__)

MASTER_HOST = "127.0.0.1"
MASTER_PORT = 9090

def send_to_master(query):
    try:
        with socket.create_connection((MASTER_HOST, MASTER_PORT), timeout=5) as sock:
            sock.sendall(query.encode('utf-8') + b'\n')
            
            if query.strip().upper().startswith('SELECT'):
                response = sock.makefile()
                
                # Read first line
                first_line = response.readline().strip()
                if first_line.startswith("ERROR:"):
                    return False, first_line[6:]
                
                if not first_line.startswith("COLUMNS:"):
                    return False, "Invalid response format"
                
                # Read columns
                num_cols = int(first_line.split(':')[1])
                columns = [response.readline().strip() for _ in range(num_cols)]
                
                # Read rows
                rows = []
                while True:
                    line = response.readline().strip()
                    if line == "END_RESULTS":
                        break
                    if line.startswith("ERROR:"):
                        return False, line[6:]
                    
                    # Read a full row
                    row = [line]
                    for _ in range(num_cols - 1):
                        row.append(response.readline().strip())
                    rows.append(row)
                
                # Format as HTML table
                table = ['<div class="table-responsive"><table class="table table-striped mt-3"><thead><tr>']
                table.extend(f'<th scope="col">{col}</th>' for col in columns)
                table.append('</tr></thead><tbody>')
                
                for row in rows:
                    table.append('<tr>')
                    table.extend(f'<td>{value}</td>' for value in row)
                    table.append('</tr>')
                
                table.append('</tbody></table></div>')
                return True, ''.join(table)
            
            elif query.strip().upper().startswith('SHOW DATABASES'):
                response = sock.makefile()
                result = response.readline().strip()
                if result.startswith("ERROR:"):
                    return False, result[6:]
                
                # Format databases in HTML
                databases = []
                while True:
                    line = response.readline().strip()
                    if line == "END_RESULTS":
                        break
                    databases.append(line)
                
                table = ['<div class="table-responsive"><table class="table table-striped mt-3"><thead><tr>']
                table.append('<th scope="col">Database Name</th>')
                table.append('</tr></thead><tbody>')
                
                for db in databases:
                    table.append(f'<tr><td>{db}</td></tr>')
                
                table.append('</tbody></table></div>')
                return True, ''.join(table)
            
            elif query.strip().upper().startswith('SHOW TABLES'):
                response = sock.makefile()
                result = response.readline().strip()
                if result.startswith("ERROR:"):
                    return False, result[6:]
                
                # Format tables in HTML
                tables = []
                while True:
                    line = response.readline().strip()
                    if line == "END_RESULTS":
                        break
                    tables.append(line)
                
                table = ['<div class="table-responsive"><table class="table table-striped mt-3"><thead><tr>']
                table.append('<th scope="col">Table Name</th>')
                table.append('</tr></thead><tbody>')
                
                for tbl in tables:
                    table.append(f'<tr><td>{tbl}</td></tr>')
                
                table.append('</tbody></table></div>')
                return True, ''.join(table)
            
            else:
                result = sock.makefile().readline().strip()
                if result.startswith("ERROR:"):
                    return False, result[6:]
                return True, "Query executed successfully"
                
    except Exception as e:
        return False, f"Connection error: {str(e)}"


HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Distributed DB Controller</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { padding: 20px; background-color: #f8f9fa; }
        .header { margin-bottom: 30px; text-align: center; }
        .query-box { min-height: 100px; font-family: monospace; }
        .result-success { background-color: #d4edda; padding: 15px; border-radius: 5px; margin-top: 20px; }
        .result-error { background-color: #f8d7da; padding: 15px; border-radius: 5px; margin-top: 20px; }
        .table-responsive { max-height: 500px; overflow-y: auto; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="text-primary">Distributed Database Controller</h1>
            <p class="text-muted">Master Node Interface</p>
        </div>
        
        <div class="card shadow">
            <div class="card-header bg-primary text-white">
                <h5>SQL Query Interface</h5>
            </div>
            <div class="card-body">
                <form action="/query" method="post">
                    <div class="mb-3">
                        <label for="sql" class="form-label">SQL Query:</label>
                        <textarea class="form-control query-box" id="sql" name="sql" 
                                placeholder="SELECT * FROM table_name; or INSERT/UPDATE/DELETE commands" 
                                required></textarea>
                    </div>
                    <button type="submit" class="btn btn-primary">Execute</button>
                    <button type="button" class="btn btn-secondary" onclick="document.getElementById('sql').value=''">Clear</button>
                </form>
                
                {% if message %}
                <div class="{% if success %}result-success{% else %}result-error{% endif %}">
                    <h5>Result:</h5>
                    {% if success and message.startswith('<div') %}
                        {{ message|safe }}
                    {% else %}
                        <pre>{{ message }}</pre>
                    {% endif %}
                    <a href="/" class="btn btn-sm btn-outline-secondary">New Query</a>
                </div>
                {% endif %}
            </div>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>
"""

@app.route('/')
def home():
    return render_template_string(HTML_TEMPLATE)

@app.route('/query', methods=['POST'])
def query():
    sql = request.form.get('sql', '').strip()
    if not sql:
        return render_template_string(HTML_TEMPLATE, 
                                  success=False, 
                                  message="Please enter a SQL query")
    
    success, message = send_to_master(sql)
    return render_template_string(HTML_TEMPLATE, 
                               success=success, 
                               message=message)

if __name__ == '__main__':
    app.run(debug=True, port=5000)