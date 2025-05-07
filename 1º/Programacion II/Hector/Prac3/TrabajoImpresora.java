package Hector.Prac3;


public class TrabajoImpresora {
    private int idTrabajador;
    private String idUsuario;
    private TrabajoImpresora siguiente;


    public TrabajoImpresora(int idTrabajador, String idUsuario){
        this.idTrabajador = idTrabajador;
        this.idUsuario = idUsuario;
    }

    public int getIdTrabajador(){
        return idTrabajador;
    }

    public String getIdUsuario(){
        return idUsuario;
    }

    public void setSiguiente(TrabajoImpresora siguiente) {
        this.siguiente = siguiente;
    }    
    
    public String toString(){
        return idTrabajador + "/" + idUsuario;
    }
}
