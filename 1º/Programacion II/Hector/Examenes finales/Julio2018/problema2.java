package Hector.Examenes finales.Julio2018;

public class problema2 {
    class TiempoUso{
        private String usuario;
        private String aplicacion;
        private int minutos;

        TiempoUso(String usuario, String aplicaion, int minutos){
            this.usuario = usuario;
            this.aplicacion = aplicaion;
            this.minutos = minutos;
        }

        String getUsuario(){
            return this.usuario;
        }

        String getAplicacion(){
            return this.aplicacion;
        }

        int getMinutos(){
            return minutos;
        }

        void aumentar(int minutos){
            minutos += minutos;
        }
    }

    class ControlParental{
        private  class Nodo {
            TiempoUso uso;
            Nodo sig;

            Nodo(TiempoUso uso, Nodo sig){
                this.uso = uso;
                this.sig = sig;
            }
        }

        private Nodo primero;

        void añadir(String usuario, String aplicacion, int minutos){
            TiempoUso nuevo = new TiempoUso(usuario, aplicacion, minutos);
            Nodo nuevoNodo = new Nodo(nuevo, null);

            if(primero == null){
                primero = nuevoNodo;
            }

            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.uso.getUsuario().equals(usuario) && actual.uso.getAplicacion().equals(aplicacion)){
                    actual.uso.minutos += minutos;
                }
            }

            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.sig == null){
                    actual.sig = nuevoNodo;
                }
            }

        }
    }
}
