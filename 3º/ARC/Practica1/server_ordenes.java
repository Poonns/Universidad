import java.util.*;
import java.io.*;
import java.net.*;

public class server_ordenes {

final static int SERVER_PORT = 8001; // puerto de escucha de nuestro servidor

public static void main (String args[]) {
Server server;
String clientRequest;
BufferedReader reader;
PrintWriter writer;

// creamos el servidor y los canales de entrada y salida

server = new Server (SERVER_PORT);
reader = new BufferedReader (new InputStreamReader (server.in));
writer = new PrintWriter (new OutputStreamWriter (server.out), true);

 

// En cuanto se establece una conexión por parte del cliente, enviamos un saludo
writer.println ("Bienvenido al Servidor: " + new Date() + "/n");

while (true) {
try {
// leemos del canal de entrada la petición del cliente
clientRequest = reader.readLine();

// Sacamos por pantalla la peticion del cliente
System.out.println ("Recibido :" + clientRequest);

// El protocolo de nuestro servidor solo acepta ordenes : HELP, QUIT,NAME,DATE
if (clientRequest.startsWith ("HELP")) {
writer.println ("Órdenes: HELP QUIT NAME DATE");
} else {
if (clientRequest.startsWith ("QUIT")) {
System.exit(0);
}
else {
if (clientRequest.startsWith("NAME"))
{InetAddress host;
host=InetAddress.getLocalHost();
writer.println("Nombre del host :"+ host.getHostName());}
else
if (clientRequest.startsWith("DATE"))
{writer.println("Fecha del sistema :"+ new Date());}



else {
writer.println ("ERROR: Comando :'" + clientRequest +"' no reconocido, use HELP");
}
}
}
} catch (IOException e) {
System.out.println ("Excepción en el servidor " + e);
System.exit(0);

}
}
}
}
//-------------------------------------------------------------------

// Esta clase es la que implementa el socket del servidor (ServerSocket)
class Server {

private ServerSocket server;
private Socket socket;

public InputStream in;
public OutputStream out;

public Server (int port) {

try {
server = new ServerSocket (port);
System.out.println ("Servidor Java Activo! \n");
System.out.println(""+server+"\n");
// Espera suspendido hasta que un cliente establece una conexión
socket = server.accept();

 

in = socket.getInputStream();
out = socket.getOutputStream();

} catch (IOException e) {
System.out.println ("Excepción en el constructor server: " + e);
}
}
} 
