//Imagina un DIOS donde un cliente pulsa un botón y la clase PuntoVenta recibe el mensaje iniciarVenta(). 
// Dentro de ese método, la línea de vida de PuntoVenta envía un mensaje de creación new() a la clase Venta y se lo guarda en su atributo interno ventaAct. 
// Escribe cómo sería el código Java aproximado del método iniciarVenta() dentro de PuntoVenta.
public class PuntoVenta{
    private Venta ventaAct

    public void iniciarVenta(){
        ventaAct = new Venta();
    }
}

private class Caja{
    private Tarjeta tarjetaCliente;

    public void procesarPago(float cantidad){
        tarjetaCliente.descontarSaldo(cantidad);
    }
}

public class Tienda{
    private BaseDatos bd;

    public Cliente antenderCliente(int idcliente){
        Cliente cliente = bd.buscarCliente(idcliente);
        Cesta nuevaCesta = new Cesta();
        cliente.asignarCesta(nuevaCesta);
    }
}

public class Tienda{
    private Almacen almacenLocal;

    public void verificarInventario(int cantidadPeticion){
        int stock = almacenLocal.consultarStock();
        if(stock < cantidadPeticion){
            almacenLocal.pedirSuministro();
        }
    }
}

public class Tienda(){
    private ArrayList<Venta> listaVentas;
    public void cerrarCaja(){
        for (Venta venta : listaVentas){
            venta.enviarAContabilidad();
        }
    }
}