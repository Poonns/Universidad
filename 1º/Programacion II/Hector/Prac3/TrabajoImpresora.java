package Hector.Prac3;


public class TrabajoImpresora {
    private int idTrabajo;
    private String idUsuario;
    private TrabajoImpresora siguiente;


    public TrabajoImpresora(int idTrabajo, String idUsuario){
        this.idTrabajo = idTrabajo;
        this.idUsuario = idUsuario;
    }

    public int getIdTrabajo(){
        return idTrabajo;
    }

    public String getIdUsuario(){
        return idUsuario;
    }

    public TrabajoImpresora getSiguiente(){
        return siguiente;
    }

    public void setSiguiente(TrabajoImpresora siguiente) {
        this.siguiente = siguiente;
    }    
    
    public String toString(){
        return idTrabajo + "/" + idUsuario;
    }
}
