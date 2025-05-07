package Hector.Prac3;

public interface ColaTrabajos {
    void insertar(TrabajoImpresora t);
    TrabajoImpresora extraerPrimero();
    TrabajoImpresora consultarPrimero();
    int getTalla();
    boolean esVacía();

}
