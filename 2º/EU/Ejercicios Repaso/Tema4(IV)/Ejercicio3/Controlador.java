public class Controlador implements KeyListener{
    private Vista vista;

    public Controlador(Vista v){
        vista = v;

        vista.getPanel().addKeyListener(new KeyAdapter(){
            @Override
            public void keyTyped(KeyEvent e){
                vista.setTecla(e.getKeyChar());
            }
        });
    }
}